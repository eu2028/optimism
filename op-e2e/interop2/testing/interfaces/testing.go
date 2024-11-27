package interfaces

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type Test interface {
	testing.TB

	Deadline() (deadline time.Time, ok bool)
	Parallel()
	Run(name string, f func(t Test)) bool
}

type WrappedT struct {
	*testing.T
}

func (t *WrappedT) Run(name string, f func(t Test)) bool {
	return t.T.Run(name, func(t *testing.T) {
		t.Helper()
		f(&WrappedT{T: t})
	})
}

func WrapT(t *testing.T) *WrappedT {
	return &WrappedT{T: t}
}

type RecoverableT struct {
	*testing.T
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

func (t *RecoverableT) Run(name string, f func(t Test)) bool {
	return t.T.Run(name, func(t *testing.T) {
		t.Helper()
		f(&RecoverableT{T: t})
	})
}

func RecoverT(t *testing.T) *RecoverableT {
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
	Setup(t Test, s S)
	Apply(t Test, s S)
}

type TestLogicFunc[S SystemBase] func(t Test, s S)

func (f TestLogicFunc[S]) Spec() TestSpec[S] {
	return &EmptyTestSpec[S]{}
}

func (f TestLogicFunc[S]) Setup(t Test, s S) {}

func (f TestLogicFunc[S]) Apply(t Test, s S) {
	f(t, s)
}
