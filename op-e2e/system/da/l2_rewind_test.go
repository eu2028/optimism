package da

import (
	"context"
	"testing"
	"time"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/stretchr/testify/require"
)

// This test verifies that the batcher can recover from a rewind of the L2 chain.
func TestBatcherL2Rewind(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	cfg.DisableProposer = true
	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	l2Seq := sys.NodeClient("sequencer")

	// Wait for a short safe chain to be built.
	safeChainLength := 5
	err = wait.ForSafeBlock(context.Background(), sys.RollupClient("sequencer"), uint64(safeChainLength))
	require.NoError(t, err)

	// rewind the L2 unsafe chain
	result := make(map[string]interface{})
	err = l2Seq.Client().Call(&result, "debug_setHead", "0x0")
	require.NoError(t, err)
	t.Log("INTERVENTION: called debugSetHead with result: ", result)

	// Wait for the intervention to take effect
	time.Sleep(time.Second * 5)

	// The batcher should be able to recover from the sequencer rewind
	require.Eventually(t, func() bool {
		ss, err := sys.RollupClient(e2esys.RoleSeq).SyncStatus(context.Background())
		if err == nil && ss.SafeL2.Number > uint64(safeChainLength) {
			return true
		}
		t.Log("Waiting for safe head to advance", "safe head", ss.SafeL2.Number)
		return false
	}, time.Second*10, time.Second, "Safe head did not advance")
}
