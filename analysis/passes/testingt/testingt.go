package testingt

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/astutil"
)

// TestingTBUnsafeFact indicates a function uses non-testing.TB methods
type TestingTBUnsafeFact struct{}

func (*TestingTBUnsafeFact) AFact() {}

func (*TestingTBUnsafeFact) String() string {
	return "TestingTBUnsafe"
}

var Analyzer = &analysis.Analyzer{
	Name:       "testingt",
	Doc:        "find constraining uses of *testing.T in non-test files",
	Run:        run,
	ResultType: nil,
	FactTypes:  []analysis.Fact{(*TestingTBUnsafeFact)(nil)},
}

func run(pass *analysis.Pass) (interface{}, error) {
	// Filter out test files first
	var nonTestFiles []*ast.File
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			nonTestFiles = append(nonTestFiles, file)
		}
	}

	// First pass: find all unsafe functions
	for _, file := range nonTestFiles {
		ast.Inspect(file, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if recv, ok := pass.TypesInfo.TypeOf(sel.X).(*types.Pointer); ok {
						if named, ok := recv.Elem().(*types.Named); ok {
							if named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "testing" && named.Obj().Name() == "T" {
								switch sel.Sel.Name {
								case "Deadline", "Run", "Parallel":
									if fn := enclosingFunction(pass, call); fn != nil {
										pass.ExportObjectFact(fn, new(TestingTBUnsafeFact))
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}

	// Second pass: find functions calling unsafe functions
	var changed bool
	for {
		changed = false
		for _, file := range nonTestFiles {
			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if fn := getFunctionObject(pass, call.Fun); fn != nil {
						if pass.ImportObjectFact(fn, new(TestingTBUnsafeFact)) {
							if caller := enclosingFunction(pass, call); caller != nil {
								if !pass.ImportObjectFact(caller, new(TestingTBUnsafeFact)) {
									pass.ExportObjectFact(caller, new(TestingTBUnsafeFact))
									changed = true
								}
							}
						}
					}
				}
				return true
			})
		}
		if !changed {
			break
		}
	}

	// Final pass: check for *testing.T usage
	for _, file := range nonTestFiles {
		ast.Inspect(file, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.Field:
				checkType(pass, n.Type, n.Pos())
			case *ast.ValueSpec:
				for _, v := range n.Values {
					checkType(pass, v, v.Pos())
				}
			}
			return true
		})
	}
	return nil, nil
}

func checkType(pass *analysis.Pass, expr ast.Expr, pos token.Pos) {
	t := pass.TypesInfo.TypeOf(expr)
	if t != nil && t.String() == "*testing.T" {
		if isTestingTBCompatible(pass, expr) {
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				Message: "avoid using *testing.T directly",
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Replace *testing.T with testing.TB",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     pos,
								End:     pos + token.Pos(len("*testing.T")),
								NewText: []byte("testing.TB"),
							},
						},
					},
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
