package runners

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers"
	"github.com/pkg/errors"
)

type SystemTest[S interfaces.SystemBase] struct {
	*testing.T
	Logic interfaces.TestLogic[S]
}

func (t SystemTest[S]) Run() {
	t.Helper()

	spec := t.Logic.Spec()
	s, err := providers.Provide[S](interfaces.WrapT(t.T), spec)
	if err != nil {
		t.Fatalf("system provider failed: %s", err)
	}
	if !spec.Conform(s) {
		t.Fatalf("system does not conform to spec")
	}

	{
		defer func() {
			if r := recover(); r != nil {
				if r, ok := r.(*interfaces.RecoverableError); ok {
					t.Fatal(errors.Wrapf(r.Err, "setup failed"))
				}
				panic(r)
			}
		}()
		t.Logic.Setup(interfaces.RecoverT(t.T), s)
	}

	t.Logic.Apply(interfaces.WrapT(t.T), s)
}
