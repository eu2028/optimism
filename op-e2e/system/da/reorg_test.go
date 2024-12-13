package da

import (
	"context"
	"math/big"
	"testing"
	"time"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/stretchr/testify/require"
)

func TestBatcherL2Rewind(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	cfg.DisableProposer = true
	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")
	l2Seq := sys.NodeClient("sequencer")

	_, err = geth.WaitForBlock(big.NewInt(12), l2Seq) // wait for 12 L2 blocks, corresponcing to 2 L1 blocks
	require.NoError(t, err, "Waiting for block on verifier")

	safeBlock, err := wait.ForNextSafeBlock(context.Background(), l2Seq)
	require.NoError(t, err)

	// rewind the L2 chain
	result := make(map[string]interface{})
	err = l2Seq.Client().Call(&result, "debug_setHead", "0x0")
	require.NoError(t, err)
	t.Log(result)

	// The batcher should be able to recover from the sequencer rewind
	require.Eventually(t, func() bool {
		ss, err := sys.RollupClient(e2esys.RoleSeq).SyncStatus(context.Background())
		if err == nil && ss.SafeL2.Number > safeBlock.Number().Uint64() {
			return true
		}
		t.Log("Waiting for safe head to advance", "safe head", ss.SafeL2.Number)
		return false
	}, time.Second*10, time.Second, "Safe head did not advance")
}

func TestBatcherL1Reorg(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	cfg.DisableProposer = true
	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")
	l2Seq := sys.NodeClient(e2esys.RoleSeq)
	l1 := sys.NodeClient(e2esys.RoleL1)

	_, err = geth.WaitForBlock(big.NewInt(12), l2Seq) // wait for 12 L2 blocks, corresponcing to 2 L1 blocks
	require.NoError(t, err, "Waiting for block on verifier")

	safeBlock, err := wait.ForNextSafeBlock(context.Background(), l2Seq)
	require.NoError(t, err)

	// reorg the L1 chain
	result := make(map[string]interface{})
	err = l1.Client().Call(&result, "debug_setHead", "0x0")
	require.NoError(t, err)
	t.Log(result)

	// The batcher should be able to recover from the gap
	// between the sequencer safe head and its oldest block.
	// We check that the safe head advances as a signal the batcher is working.
	require.Eventually(t, func() bool {
		ss, err := sys.RollupClient(e2esys.RoleSeq).SyncStatus(context.Background())
		if err == nil && ss.SafeL2.Number > safeBlock.NumberU64() {
			return true
		}
		t.Log("Waiting for safe head to advance", "safe head", ss.SafeL2.Number)
		return false
	}, time.Second*30, time.Second, "Safe head did not advance")
}
