package interop2

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/providers/e2e_backends"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	gethCore "github.com/ethereum/go-ethereum/core"
	gethTypes "github.com/ethereum/go-ethereum/core/types"
)

func testInteropNoop(t Test, s SuperSystem) {
	t.Helper()
}

func TestInteropNoop(t *testing.T) {
	SystemTest{T: t, Logic: TestLogicFunc(testInteropNoop)}.Run()
}

type behaviorModel struct {
	invalidPayloadExpectedError        error
	invalidPayloadExecutionExpectation func(context.Context, Test, error)
	noError                            func(context.Context, Test, error)
}

func getBehaviorModel(mempoolFiltering bool) *behaviorModel {
	model := &behaviorModel{
		noError: func(ctx context.Context, t Test, err error) {
			require.NoError(t, err)
		},
	}

	if mempoolFiltering {
		model.invalidPayloadExpectedError = gethCore.ErrTxFilteredOut
	} else {
		model.invalidPayloadExpectedError = nil
	}

	model.invalidPayloadExecutionExpectation = func(ctx context.Context, t Test, err error) {
		if mempoolFiltering {
			require.ErrorContains(t, err, gethCore.ErrTxFilteredOut.Error())
		} else {
			require.ErrorIs(t, err, ctx.Err())
			require.ErrorIs(t, ctx.Err(), context.DeadlineExceeded)
		}
	}
	return model
}

type testInteropBlockBuilding struct {
	spec *e2e_backends.SuperSystemSpec

	alice string
	bob   string

	chainA string
	chainB string

	identifier types.Identifier
	ev         *gethTypes.Log
}

func (ti *testInteropBlockBuilding) Spec() TestSpec {
	// TODO: we should push the N required chains into the spec
	return ti.spec
}

func randomizeUser(name string) string {
	return fmt.Sprintf("%s_%d", name, time.Now().UnixNano())
}

func (ti *testInteropBlockBuilding) Setup(t Test, s SuperSystem) error {
	logger := testlog.Logger(t, log.LevelInfo)
	// oplog.SetGlobalLogHandler(logger.Handler())

	// TODO: scope to test instance
	ti.alice = randomizeUser("Alice")
	ti.bob = randomizeUser("Bob")
	s.AddUser(ti.alice)
	s.AddUser(ti.bob)

	// TODO: break this into smaller operations
	ids := s.L2IDs()
	ti.chainA = ids[0]
	ti.chainB = ids[1]

	// We will initiate on chain A, and execute on chain B
	s.DeployEmitterContract(ti.chainA, ti.alice)

	// Add chain A as dependency to chain B,
	// such that we can execute a message on B that was initiated on A.
	depRec := s.AddDependency(ti.chainB, s.ChainID(ti.chainA))

	rollupClA, err := dial.DialRollupClientWithTimeout(context.Background(), time.Second*15, logger, s.OpNode(ti.chainA).UserRPC().RPC())
	if err != nil {
		return err
	}

	// Now wait for the dependency to be visible in the L2 (receipt needs to be picked up)
	require.Eventually(t, func() bool {
		status, err := rollupClA.SyncStatus(context.Background())
		require.NoError(t, err)
		return status.CrossUnsafeL2.L1Origin.Number >= depRec.BlockNumber.Uint64()
	}, time.Second*30, time.Second, "wait for L1 origin to match dependency L1 block")
	t.Log("Dependency information has been processed in L2 block")

	// emit log on chain A
	emitRec := s.EmitData(ti.chainA, ti.alice, "hello world")
	t.Logf("Emitted a log event in block %d", emitRec.BlockNumber.Uint64())

	// Wait for initiating side to become cross-unsafe
	require.Eventually(t, func() bool {
		status, err := rollupClA.SyncStatus(context.Background())
		require.NoError(t, err)
		return status.CrossUnsafeL2.Number >= emitRec.BlockNumber.Uint64()
	}, time.Second*60, time.Second, "wait for emitted data to become cross-unsafe")
	t.Logf("Reached cross-unsafe block %d", emitRec.BlockNumber.Uint64())

	// Identify the log
	require.Len(t, emitRec.Logs, 1)
	ev := emitRec.Logs[0]
	ethCl := s.L2GethClient(ti.chainA)
	header, err := ethCl.HeaderByHash(context.Background(), emitRec.BlockHash)
	require.NoError(t, err)

	ti.ev = ev
	ti.identifier = types.Identifier{
		Origin:      ev.Address,
		BlockNumber: ev.BlockNumber,
		LogIndex:    uint32(ev.Index),
		Timestamp:   header.Time,
		ChainID:     types.ChainIDFromBig(s.ChainID(ti.chainA)),
	}

	return nil
}

func (ti *testInteropBlockBuilding) Apply(t Test, s SuperSystem) {
	model := getBehaviorModel(ti.spec.Config.MempoolFiltering)
	data := []struct {
		name                 string
		expectedError        error
		payload              []byte
		executionExpectation func(context.Context, Test, error)
	}{
		{
			name:                 "invalid message",
			payload:              []byte("test invalid message"),
			expectedError:        model.invalidPayloadExpectedError,
			executionExpectation: model.invalidPayloadExecutionExpectation,
		},
		{
			name:                 "valid message",
			payload:              types.LogToMessagePayload(ti.ev),
			expectedError:        nil,
			executionExpectation: model.noError,
		},
	}

	for _, tt := range data {
		t.Run(tt.name, func(t Test) {
			bobAddr := s.Address(ti.chainA, ti.bob) // direct it to a random account without code
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
			defer cancel()

			_, err := s.ExecuteMessage(ctx, ti.chainB, ti.alice, ti.identifier, bobAddr, tt.payload, tt.expectedError)
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
		t.Run(fmt.Sprintf("mempool_filtering=%t", mempoolFiltering), func(t *testing.T) {
			t.Parallel()
			spec := &e2e_backends.SuperSystemSpec{
				Config: e2e_backends.SuperSystemConfig{
					MempoolFiltering: mempoolFiltering,
				},
			}
			SystemTest{T: t, Logic: &testInteropBlockBuilding{spec: spec}}.Run()
		})
	}
}
