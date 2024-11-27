package interop2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers/e2e_backends"
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
	setupSyncPoint *e2e_backends.SyncPoint
	users          []string
	chains         []string
}

func (ti *testInteropBlockBuilding) Spec() TestSpec {
	// TODO: we should push the N required chains into the spec ?
	return ti.spec
}

func (ti *testInteropBlockBuilding) Setup(t Test, s SuperSystem) error {
	auto := &e2e_backends.SuperSystemAutomation{
		Sys:    s,
		Logger: testlog.Logger(t, log.LevelInfo),
		T:      t,
	}
	// oplog.SetGlobalLogHandler(logger.Handler())

	ti.users = make([]string, numberOfUsers)
	for i := 0; i < numberOfUsers; i++ {
		ti.users[i] = auto.NewUniqueUser(fmt.Sprintf("User%d", i))
	}

	ti.chains = s.L2IDs()

	err := auto.SetupXChainMessaging(ti.users[0], ti.chains[0], ti.chains[1])
	require.NoError(t, err)

	// emit log on chain A
	syncPoint, err := auto.SendXChainMessage(ti.users[0], ti.chains[0], "hello world")
	require.NoError(t, err)

	ti.setupSyncPoint = syncPoint
	return nil
}

func (ti *testInteropBlockBuilding) Apply(t Test, s SuperSystem) {
	model := helpers.GetBehaviorModel(ti.spec.Config.MempoolFiltering())
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
			bobAddr := s.Address(ti.chains[0], ti.users[1]) // direct it to a random account without code
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			_, err := s.ExecuteMessage(ctx, ti.chains[1], ti.users[0], ti.setupSyncPoint.Identifier(), bobAddr, tt.payload, tt.expectedError)
			tt.executionExpectation(ctx, t, err)
		})
	}
}

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
