package interop2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/automation"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/expectations"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

const (
	numberOfL2s   = 2
	numberOfUsers = 2
)

func testInteropNoop(t Test, s SuperSystem) {
	t.Helper()
}

// TestInteropNoop is a test that does nothing but bring up a stack.
func TestInteropNoop(t *testing.T) {
	SystemTest{T: t, Logic: TestLogicFunc(testInteropNoop)}.Run()
}

type testInteropBlockBuilding struct {
	spec           *interfaces.SuperSystemSpec
	setupSyncPoint *automation.SyncPoint
	auto           *automation.SuperSystemAutomation
}

func (ti *testInteropBlockBuilding) Spec() TestSpec {
	return ti.spec
}

func (ti *testInteropBlockBuilding) getShorthands() (alice, bob, chainA, chainB string) {
	alice = ti.auto.User(0)
	bob = ti.auto.User(1)
	chainA = ti.auto.Chain(0)
	chainB = ti.auto.Chain(1)
	return
}

func (ti *testInteropBlockBuilding) Setup(t Test, s SuperSystem) {
	ti.auto = automation.NewSuperSystemAutomation(s, testlog.Logger(t, log.LevelInfo), t)
	// oplog.SetGlobalLogHandler(logger.Handler())
	ti.auto.NewUniqueUsers(numberOfUsers)

	alice, _, chainA, chainB := ti.getShorthands()

	err := ti.auto.SetupXChainMessaging(alice, chainA, chainB)
	require.NoError(t, err)

	// emit log on chain A
	syncPoint, err := ti.auto.SendXChainMessage(alice, chainA, "hello world")
	require.NoError(t, err)

	ti.setupSyncPoint = syncPoint
}

func (ti *testInteropBlockBuilding) Apply(t Test, s SuperSystem) {
	model := expectations.GetBehaviorModel(ti.spec.Config.MempoolFiltering())
	alice, bob, chainA, chainB := ti.getShorthands()

	data := []struct {
		name                 string
		expectedError        error
		payload              []byte
		executionExpectation func(context.Context, Test, error)
	}{
		{
			name:                 "invalid message",
			payload:              []byte("test invalid message"),
			expectedError:        model.InvalidPayloadExpectedError,
			executionExpectation: model.InvalidPayloadExecutionExpectation,
		},
		{
			name:                 "valid message",
			payload:              types.LogToMessagePayload(ti.setupSyncPoint.Event()),
			expectedError:        nil,
			executionExpectation: model.NoError,
		},
	}

	for _, tt := range data {
		t.Run(tt.name, func(t Test) {
			bobAddr := s.Address(chainA, bob) // direct it to a random account without code
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			_, err := s.ExecuteMessage(ctx, chainB, alice, ti.setupSyncPoint.Identifier(), bobAddr, tt.payload, tt.expectedError)
			tt.executionExpectation(ctx, t, err)
		})
	}
}

var _ TestLogic = (*testInteropBlockBuilding)(nil)

func TestInteropBlockBuilding(t *testing.T) {
	for _, useFiltering := range []bool{
		false,
		true,
	} {
		mempoolFiltering := useFiltering
		t.Run(fmt.Sprintf("mempoolFiltering=%t", mempoolFiltering), func(t *testing.T) {
			t.Parallel()
			spec := &interfaces.SuperSystemSpec{
				Config: interfaces.NewSuperSystemConfig(
					interfaces.WithMempoolFiltering(mempoolFiltering),
					interfaces.WithNumberOfL2s(numberOfL2s),
				),
			}
			SystemTest{T: t, Logic: &testInteropBlockBuilding{
				spec: spec,
			}}.Run()
		})
	}
}
