package interfaces

import (
	"testing"
	"time"
)

type Test interface {
	testing.TB

	Deadline() (deadline time.Time, ok bool)
	Parallel()
	Run(name string, f func(t Test)) bool
}

// TODO: add support for testing phases.
// to separate setup failures from apply failures.
type WrappedT struct {
	*testing.T
}

func (t *WrappedT) Deadline() (deadline time.Time, ok bool) {
	return t.T.Deadline()
}

func (t *WrappedT) Run(name string, f func(t Test)) bool {
	return t.T.Run(name, func(t *testing.T) {
		t.Helper()
		f(&WrappedT{T: t})
	})
}

func (t *WrappedT) Parallel() {
	t.T.Parallel()
}

func WrapT(t *testing.T) *WrappedT {
	return &WrappedT{T: t}
}

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
