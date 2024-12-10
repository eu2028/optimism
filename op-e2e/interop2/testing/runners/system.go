package runners

import (
	"fmt"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/sigma/go-test-trace/pkg/trace_testing"
)

type SystemTest[S interfaces.SystemBase] struct {
	trace_testing.T
	Logic interfaces.TestLogic[S]
}

func recoverPhase(t trace_testing.T, tracer trace.Tracer, phase string) (trace_testing.T, func()) {
	ctx, span := tracer.Start(t.Context(), phase)
	return t.WithContext(ctx), func() {
		defer span.End()
		if r := recover(); r != nil {
			if r, ok := r.(*interfaces.RecoverableError); ok {
				msg := fmt.Sprintf("%s phase failed", phase)
				t.Fatal(errors.Wrap(r.Err, msg))
			}
		}
	}
}

func (t SystemTest[S]) Run() {
	tracer := otel.Tracer("system test")

	t.Helper()

	spec := t.Logic.Spec()
	var system S

	phases := []struct {
		name string
		fn   func(t trace_testing.T)
	}{
		{
			name: "system provisioning",
			fn: func(tt trace_testing.T) {
				s, err := providers.Provide[S](interfaces.RecoverT(tt), spec)
				if err != nil {
					tt.Fatalf("provider failed: %s", err)
				}
				system = s
			},
		},
		{
			name: "spec conformance",
			fn: func(tt trace_testing.T) {
				if !spec.Conform(system) {
					tt.Fatalf("system does not conform to spec")
				}
			},
		},
		{
			name: "SUT setup",
			fn: func(tt trace_testing.T) {
				if logic, ok := t.Logic.(interfaces.TestLogicSetup[S]); ok {
					logic.Setup(interfaces.RecoverT(tt), system)
				}
			},
		},
		{
			name: "test execution",
			fn: func(tt trace_testing.T) {
				t.Logic.Check(tt, system)
			},
		},
		{
			name: "SUT cleanup",
			fn: func(tt trace_testing.T) {
				if logic, ok := t.Logic.(interfaces.TestLogicCleanup[S]); ok {
					logic.Cleanup(interfaces.RecoverT(tt), system)
				}
			},
		},
	}

	for _, phase := range phases {
		func() {
			tt, recover := recoverPhase(t.T, tracer, phase.name)
			defer recover()
			phase.fn(tt)
		}()
	}
}
