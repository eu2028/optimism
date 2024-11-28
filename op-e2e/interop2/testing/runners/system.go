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
				msg := fmt.Sprintf("%s phase failed", phase)
				t.Fatal(errors.Wrap(r.Err, msg))
			}
		}
	}
}

func (t SystemTest[S]) Run() {
	t.Helper()

	spec := t.Logic.Spec()
	var system S

	{
		defer recoverPhase(t, "system provisioning")()
		s, err := providers.Provide[S](interfaces.RecoverT(t.T), spec)
		if err != nil {
			t.Fatalf("provider failed: %s", err)
		}
		system = s
	}

	{
		defer recoverPhase(t, "spec conformance")()
		if !spec.Conform(system) {
			t.Fatalf("system does not conform to spec")
		}
	}

	if logic, ok := t.Logic.(interfaces.TestLogicSetup[S]); ok {
		defer recoverPhase(t, "SUT setup")()
		logic.Setup(interfaces.RecoverT(t.T), system)
	}

	{
		defer recoverPhase(t, "test execution")()
		t.Logic.Check(interfaces.WrapT(t.T), system)
	}

	if logic, ok := t.Logic.(interfaces.TestLogicCleanup[S]); ok {
		defer recoverPhase(t, "SUT cleanup")()
		logic.Cleanup(interfaces.RecoverT(t.T), system)
	}
}
