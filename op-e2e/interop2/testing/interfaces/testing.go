package interfaces

import (
	"errors"
	"fmt"

	"github.com/sigma/go-test-trace/pkg/trace_testing"
)

type Test = trace_testing.T

type RecoverableT struct {
	trace_testing.T
}

type RecoverableError struct {
	Err error
}

func (t *RecoverableT) FailNow() {
	panic(&RecoverableError{
		Err: errors.New("failed"),
	})
}

func (t *RecoverableT) Fatal(args ...any) {
	t.T.Log(args...)
	t.FailNow()
}

func (t *RecoverableT) Fatalf(format string, args ...any) {
	t.T.Logf(format, args...)
	panic(&RecoverableError{
		Err: fmt.Errorf(format, args...),
	})
}

func (t *RecoverableT) SkipNow() {
	panic(&RecoverableError{
		Err: errors.New("skipped"),
	})
}

func RecoverT(t Test) *RecoverableT {
	return &RecoverableT{T: t}
}

var _ Test = (*RecoverableT)(nil)

type SystemBase = any

type TestSpec[S SystemBase] interface {
	Conform(s S) bool
}

type EmptyTestSpec[S SystemBase] struct{}

func (EmptyTestSpec[S]) Conform(s S) bool {
	return true
}

type TestLogic[S SystemBase] interface {
	Spec() TestSpec[S]
	Check(t Test, s S)
}

type TestLogicSetup[S SystemBase] interface {
	Setup(t Test, s S)
}

type TestLogicCleanup[S SystemBase] interface {
	Cleanup(t Test, s S)
}

type TestLogicFunc[S SystemBase] func(t Test, s S)

func (f TestLogicFunc[S]) Spec() TestSpec[S] {
	return &EmptyTestSpec[S]{}
}

func (f TestLogicFunc[S]) Check(t Test, s S) {
	f(t, s)
}

var _ TestLogic[SystemBase] = TestLogicFunc[SystemBase](nil)
