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
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

// TestSystemBatchType* run each system e2e test case in singular batch mode and span batch mode.
// If the test case tests batch submission and advancing safe head, it should be tested in both singular and span batch mode.
func TestSystemBatchType_SingularBatch(t *testing.T) {
	testStartStopBatcher(t, func(sc *e2esys.SystemConfig) {
		sc.BatcherBatchType = derive.SingularBatchType
	})
}

func TestSystemBatchType_SpanBatch(t *testing.T) {
	testStartStopBatcher(t, func(sc *e2esys.SystemConfig) {
		sc.BatcherBatchType = derive.SpanBatchType
	})
}

func TestSystemBatchType_SpanBatchMaxBlocks(t *testing.T) {
	testStartStopBatcher(t, func(sc *e2esys.SystemConfig) {
		sc.BatcherBatchType = derive.SpanBatchType
		sc.BatcherMaxBlocksPerSpanBatch = 2
	})
}

func testStartStopBatcher(t *testing.T, cfgMod func(*e2esys.SystemConfig)) {
	op_e2e.InitParallel(t)

	cfg := e2esys.DefaultSystemConfig(t)
	cfgMod(&cfg)
	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	rollupClient := sys.RollupClient("verifier")

	l2Seq := sys.NodeClient("sequencer")
	l2Verif := sys.NodeClient("verifier")

	// retrieve the initial sync status
	seqStatus, err := rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)

	nonce := uint64(0)
	sendTx := func() *types.Receipt {
		// Submit TX to L2 sequencer node
		receipt := helpers.SendL2Tx(t, cfg, l2Seq, cfg.Secrets.Alice, func(opts *helpers.TxOpts) {
			opts.ToAddr = &common.Address{0xff, 0xff}
			opts.Value = big.NewInt(1_000_000_000)
			opts.Nonce = nonce
		})
		nonce++
		return receipt
	}
	// send a transaction
	receipt := sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the verifier
	safeBlockInclusionDuration := time.Duration(6*cfg.DeployConfig.L1BlockTime) * time.Second
	_, err = geth.WaitForBlock(receipt.BlockNumber, l2Verif)
	require.NoError(t, err, "Waiting for block on verifier")
	require.NoError(t, wait.ForProcessingFullBatch(context.Background(), rollupClient))

	// ensure the safe chain advances
	newSeqStatus, err := rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance")

	driver := sys.BatchSubmitter.TestDriver()
	// stop the batch submission
	err = driver.StopBatchSubmitting(context.Background())
	require.NoError(t, err)

	// wait for any old safe blocks being submitted / derived
	time.Sleep(safeBlockInclusionDuration)

	// get the initial sync status
	seqStatus, err = rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)

	// send another tx
	sendTx()
	time.Sleep(safeBlockInclusionDuration)

	// ensure that the safe chain does not advance while the batcher is stopped
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)
	require.Equal(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain advanced while batcher was stopped")

	// start the batch submission
	err = driver.StartBatchSubmitting()
	require.NoError(t, err)
	time.Sleep(safeBlockInclusionDuration)

	// send a third tx
	receipt = sendTx()

	// wait until the block the tx was first included in shows up in the safe chain on the verifier
	_, err = geth.WaitForBlock(receipt.BlockNumber, l2Verif)
	require.NoError(t, err, "Waiting for block on verifier")
	require.NoError(t, wait.ForProcessingFullBatch(context.Background(), rollupClient))

	// ensure that the safe chain advances after restarting the batcher
	newSeqStatus, err = rollupClient.SyncStatus(context.Background())
	require.NoError(t, err)
	require.Greater(t, newSeqStatus.SafeL2.Number, seqStatus.SafeL2.Number, "Safe chain did not advance after batcher was restarted")
}

// TestBatcherSafeHeadGap tests that the batcher can handle
// an (effective) reversal in the sequencer's safe head.
// This can happen when the sequencer restarts, e.g. during a rollout.
// To simulate the reversal of the safe head, we instead set a single
// block in the batcher's state with a huge block number.
func TestBatcherSafeHeadGap(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	sys, err := cfg.Start(t)
	require.NoError(t, err, "Error starting up system")

	l2Seq := sys.NodeClient("sequencer")

	nonce := uint64(0)
	sendTx := func() *types.Receipt {
		// Submit TX to L2 sequencer node
		receipt := helpers.SendL2Tx(t, cfg, l2Seq, cfg.Secrets.Alice, func(opts *helpers.TxOpts) {
			opts.ToAddr = &common.Address{0xff, 0xff}
			opts.Value = big.NewInt(1_000_000_000)
			opts.Nonce = nonce
		})
		nonce++
		return receipt
	}

	// send some transactions
	for i := 0; i < 3; i++ {
		sendTx()
	}

	sys.BatchSubmitter.TestDriver().ForceOldestBlockIntoFuture()
	t.Log("Set batcher's oldest block number to MaxInt64")

	block, err := wait.ForNextSafeBlock(context.Background(), l2Seq)
	require.NoError(t, err)

	// send some transactions
	for i := 0; i < 3; i++ {
		sendTx()
	}

	// The batcher should be able to recover from the gap
	// between the sequencer safe head and its oldest block.
	// We check that the safe head advances as a signal the batcher is working.
	require.Eventually(t, func() bool {
		ss, err := sys.RollupClient(e2esys.RoleSeq).SyncStatus(context.Background())
		if err == nil && ss.SafeL2.Number > block.Number().Uint64() {
			return true
		}
		return false
	}, time.Second*10, time.Second, "Safe head did not advance")
}
