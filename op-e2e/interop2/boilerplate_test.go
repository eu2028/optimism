package interop2

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/runners"
	"github.com/sigma/go-test-trace/pkg/trace_testing"
)

type SuperSystem = interfaces.SuperSystem
type Test = interfaces.Test
type TestSpec = interfaces.TestSpec[SuperSystem]
type SystemTest = runners.SystemTest[SuperSystem]
type TestLogic = interfaces.TestLogic[SuperSystem]
type TestLogicFunc = interfaces.TestLogicFunc[SuperSystem]

type T = trace_testing.T

func WithExtendedT(t *testing.T) T {
	return trace_testing.WithTracing(t)
}
