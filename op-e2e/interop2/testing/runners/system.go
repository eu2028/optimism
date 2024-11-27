package runners

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers"
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
		t.Fatalf("failed to provide system: %s", err)
	}
	if !spec.Conform(s) {
		t.Fatalf("system does not conform to spec")
	}

	wrapped := interfaces.WrapT(t.T)
	t.Logic.Setup(wrapped, s)
	t.Logic.Apply(wrapped, s)
}
