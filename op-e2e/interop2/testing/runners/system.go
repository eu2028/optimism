package runners

import (
	"fmt"
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers"
	"github.com/pkg/errors"
)

type SystemTest[S interfaces.SystemBase] struct {
	*testing.T
	Logic interfaces.TestLogic[S]
}

func recoverPhase(t testing.TB, phase string) func() {
	return func() {
		if r := recover(); r != nil {
			if r, ok := r.(*interfaces.RecoverableError); ok {
				msg := fmt.Sprintf("%s failed", phase)
				t.Fatal(errors.Wrapf(r.Err, msg))
			}
		}
	}
}

func (t SystemTest[S]) Run() {
	t.Helper()

	spec := t.Logic.Spec()
	var system S

	{
		defer recoverPhase(t, "system provider")()

		s, err := providers.Provide[S](interfaces.RecoverT(t.T), spec)
		if err != nil {
			t.Fatalf("system provider failed: %s", err)
		}
		if !spec.Conform(s) {
			t.Fatalf("system does not conform to spec")
		}
		system = s
	}

	{
		defer recoverPhase(t, "setup")()
		t.Logic.Setup(interfaces.RecoverT(t.T), system)
	}

	t.Logic.Apply(interfaces.WrapT(t.T), system)
}
