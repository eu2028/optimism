// Package testingt finds uses of *testing.T in non-test files.
//
// Those are a testability issue as *testing.T cannot be constructed. So the
// corresponding code can only run in a proper test. Generally there's no good
// reason to tie the implementation to testing infrastructure though. Using
// testing.TB is more flexible, and opens the possibility of running the code
// outside of tests (or with tests that use a different testing framework).
package testingt

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
)

// TestingTBUnsafeFact indicates a function uses non-testing.TB methods
type TestingTBUnsafeFact struct{}

func (*TestingTBUnsafeFact) AFact() {}

func (*TestingTBUnsafeFact) String() string {
	return "TestingTBUnsafe"
}

// unsafeTestingTMethods defines methods that are specific to *testing.T and cannot
// be used with testing.TB interface
var unsafeTestingTMethods = map[string]bool{
	"Deadline": true,
	"Run":      true,
	"Parallel": true,
}

var Analyzer = &analysis.Analyzer{
	Name:       "testingt",
	Doc:        "find constraining uses of *testing.T in non-test files",
	Run:        run,
	Requires:   []*analysis.Analyzer{inspect.Analyzer},
	ResultType: nil,
	FactTypes:  []analysis.Fact{(*TestingTBUnsafeFact)(nil)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Filter out test files first
	var nonTestFiles []*ast.File
	inspect.Preorder([]ast.Node{(*ast.File)(nil)}, func(n ast.Node) {
		file := n.(*ast.File)
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			nonTestFiles = append(nonTestFiles, file)
		}
	})

	// First pass: find all unsafe functions
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		if !isInNonTestFile(n, nonTestFiles) {
			return
		}
		call := n.(*ast.CallExpr)
		if isUnsafeTestingTMethod(pass, call) {
			if fn := enclosingFunction(pass, call); fn != nil {
				pass.ExportObjectFact(fn, new(TestingTBUnsafeFact))
			}
		}
	})

	// Second pass: find functions calling unsafe functions
	var changed bool
	for {
		changed = false
		inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
			if !isInNonTestFile(n, nonTestFiles) {
				return
			}
			call := n.(*ast.CallExpr)
			changed = checkAndMarkCallerUnsafe(pass, call) || changed
		})
		if !changed {
			break
		}
	}

	// Final pass: check for *testing.T usage
	inspect.Preorder([]ast.Node{(*ast.Field)(nil), (*ast.ValueSpec)(nil)}, func(n ast.Node) {
		if !isInNonTestFile(n, nonTestFiles) {
			return
		}
		switch n := n.(type) {
		case *ast.Field:
			checkType(pass, n.Type, n.Pos())
		case *ast.ValueSpec:
			for _, v := range n.Values {
				checkType(pass, v, v.Pos())
			}
		}
	})

	return nil, nil
}

func isInNonTestFile(n ast.Node, nonTestFiles []*ast.File) bool {
	pos := n.Pos()
	for _, f := range nonTestFiles {
		if f.Pos() <= pos && pos <= f.End() {
			return true
		}
	}
	return false
}

func checkType(pass *analysis.Pass, expr ast.Expr, pos token.Pos) {
	t := pass.TypesInfo.TypeOf(expr)
	if t != nil && t.String() == "*testing.T" {
		if isTestingTBCompatible(pass, expr) {
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: "avoid using *testing.T directly",
				SuggestedFixes: []analysis.SuggestedFix{
					createTestingTBSuggestedFix(pos),
				},
			})
		} else {
			pass.Reportf(pos, "avoid using *testing.T directly")
		}
	}
}

func isTestingTBCompatible(pass *analysis.Pass, expr ast.Expr) bool {
	// Check if the expression is used in any unsafe function
	var isUnsafe bool
	ast.Inspect(expr, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if fn := getFunctionObject(pass, call.Fun); fn != nil {
				if pass.ImportObjectFact(fn, new(TestingTBUnsafeFact)) {
					isUnsafe = true
					return false
				}
			}
		}
		return true
	})
	return !isUnsafe
}

func enclosingFunction(pass *analysis.Pass, node ast.Node) *types.Func {
	var file *ast.File
	pos := node.Pos()
	for _, f := range pass.Files {
		if f.Pos() <= pos && pos <= f.End() {
			file = f
			break
		}
	}
	if file == nil {
		return nil
	}
	path, _ := astutil.PathEnclosingInterval(file, node.Pos(), node.End())
	for _, n := range path {
		if fn, ok := n.(*ast.FuncDecl); ok {
			return pass.TypesInfo.ObjectOf(fn.Name).(*types.Func)
		}
	}
	return nil
}

func getFunctionObject(pass *analysis.Pass, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return pass.TypesInfo.ObjectOf(expr)
	case *ast.SelectorExpr:
		return pass.TypesInfo.ObjectOf(expr.Sel)
	}
	return nil
}

// isUnsafeTestingTMethod checks if the call expression is using a *testing.T-specific method
func isUnsafeTestingTMethod(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	recv, ok := pass.TypesInfo.TypeOf(sel.X).(*types.Pointer)
	if !ok {
		return false
	}

	named, ok := recv.Elem().(*types.Named)
	if !ok {
		return false
	}

	return named.Obj().Pkg() != nil &&
		named.Obj().Pkg().Path() == "testing" &&
		named.Obj().Name() == "T" &&
		unsafeTestingTMethods[sel.Sel.Name]
}

func checkAndMarkCallerUnsafe(pass *analysis.Pass, call *ast.CallExpr) bool {
	fn := getFunctionObject(pass, call.Fun)
	if fn == nil {
		return false
	}

	if !pass.ImportObjectFact(fn, new(TestingTBUnsafeFact)) {
		return false
	}

	caller := enclosingFunction(pass, call)
	if caller == nil {
		return false
	}

	if !pass.ImportObjectFact(caller, new(TestingTBUnsafeFact)) {
		pass.ExportObjectFact(caller, new(TestingTBUnsafeFact))
		return true
	}

	return false
}

func createTestingTBSuggestedFix(pos token.Pos) analysis.SuggestedFix {
	return analysis.SuggestedFix{
		Message: "Replace *testing.T with testing.TB",
		TextEdits: []analysis.TextEdit{
			{
				Pos:     pos,
				End:     pos + token.Pos(len("*testing.T")),
				NewText: []byte("testing.TB"),
			},
		},
	}
}
