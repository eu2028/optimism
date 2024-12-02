package interop

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/interop/contracts/bindings/emit"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/interop/contracts/bindings/inbox"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/interop/contracts/bindings/systemconfig"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
)

// func TestInteropMessageExecution(gt *testing.T) {
// 	t := helpers.NewDefaultTesting(gt)
// 	is := SetupInterop(t)
// 	actors := is.CreateActors()
//
// 	chainAUserKeys := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)
// 	chainBUserKeys := devkeys.ChainUserKeys(actors.ChainB.ChainCfg.ChainID)
//
// 	userAChainAKey := chainAUserKeys(0)
// 	userAChainBKey := chainBUserKeys(0)
//
// 	// Get both sequencers ready
// 	actors.ChainA.Sequencer.ActL2PipelineFull(t)
// 	actors.ChainB.Sequencer.ActL2PipelineFull(t)
//
// 	// We'll use User A for our testing
// 	secret, err := is.Keys.Secret(userAChainAKey)
// 	require.NoError(t, err)
//
// 	// Create transaction options for the user
// 	deployerAuth, err := bind.NewKeyedTransactorWithChainID(
// 		secret,
// 		actors.ChainA.ChainCfg.ChainID,
// 	)
// 	require.NoError(t, err)
// 	deployerAuth.GasLimit = 3000000
// 	deployerAuth.GasTipCap = big.NewInt(2 * params.GWei)
//
// 	// Deploy the Emit contract on Chain A
// 	emitAddr, tx, emitContract, err := emit.DeployEmit(deployerAuth, actors.ChainA.SequencerEngine.EthClient())
// 	require.NoError(t, err)
//
// 	actors.ChainA.Sequencer.ActL2StartBlock(t)
// 	err = actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(secret.PublicKey))
// 	require.NoError(t, err)
// 	actors.ChainA.Sequencer.ActL2EndBlock(t)
//
// 	// Create test user
// 	userSecret, err := is.Keys.Secret(userAChainAKey)
// 	require.NoError(t, err)
// 	userAuthA, err := bind.NewKeyedTransactorWithChainID(
// 		userSecret,
// 		actors.ChainA.ChainCfg.ChainID,
// 	)
// 	require.NoError(t, err)
//
// 	// Create and emit the test message
// 	testData := []byte("hello chain B!")
// 	tx, err = emitContract.EmitData(userAuthA, testData)
// 	require.NoError(t, err)
//
// 	// Include the emit transaction
// 	actors.ChainA.Sequencer.ActL2StartBlock(t)
// 	err = actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(secret.PublicKey))
// 	require.NoError(t, err)
// 	actors.ChainA.Sequencer.ActL2EndBlock(t)
//
// 	// Submit batch and mine it
// 	actors.ChainA.Batcher.ActSubmitAll(t)
// 	actors.L1Miner.ActL1StartBlock(12)(t)
// 	actors.L1Miner.ActL1IncludeTx(actors.ChainA.BatcherAddr)(t)
// 	actors.L1Miner.ActL1EndBlock(t)
//
// 	// Sync Chain A status to supervisor and wait for safety
// 	actors.Supervisor.SyncEvents(t, actors.ChainA.ChainID)
// 	actors.ChainA.Sequencer.ActL2PipelineFull(t)
// 	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainA.ChainID)
// 	actors.Supervisor.SyncCrossSafe(t, actors.ChainA.ChainID)
//
// 	// Now try to execute the message on Chain B using the same user
// 	userSecret, err = is.Keys.Secret(userAChainBKey)
// 	require.NoError(t, err)
// 	userAuthB, err := bind.NewKeyedTransactorWithChainID(
// 		userSecret,
// 		actors.ChainB.ChainCfg.ChainID,
// 	)
// 	require.NoError(t, err)
// 	userAuthB.GasLimit = 3000000
// 	userAuthB.GasTipCap = big.NewInt(2 * params.GWei)
//
// 	inboxContract, err := inbox.NewInbox(predeploys.CrossL2InboxAddr, actors.ChainB.SequencerEngine.EthClient())
// 	require.NoError(t, err)
//
// 	// Create the ExecuteMessage tx for chain B
// 	status := actors.ChainB.Sequencer.SyncStatus()
// 	identifier := inbox.Identifier{
// 		Origin:      emitAddr,
// 		BlockNumber: big.NewInt(2), // TODO: Get actual block number
// 		LogIndex:    big.NewInt(0),
// 		Timestamp:   big.NewInt(int64(status.UnsafeL2.Time)),
// 		ChainId:     actors.ChainA.ChainCfg.ChainID,
// 	}
//
// 	tx, err = inboxContract.ExecuteMessage(userAuthB, identifier, emitAddr, testData)
// 	require.NoError(t, err)
//
// 	// Include the ExecuteMessage tx on Cha
// 	actors.ChainB.Sequencer.ActL2StartBlock(t)
// 	err = actors.ChainB.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
// 	require.NoError(t, err)
// 	actors.ChainB.Sequencer.ActL2EndBlock(t)
//
// 	// Sync Chain B and verify message execution
// 	actors.Supervisor.SyncEvents(t, actors.ChainB.ChainID)
// 	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainB.ChainID)
// 	actors.Supervisor.SyncCrossSafe(t, actors.ChainB.ChainID)
// }

// func TestInteropMessageExecution(gt *testing.T) {
// 	t := helpers.NewDefaultTesting(gt)
// 	is := SetupInterop(t)
// 	actors := is.CreateActors()
//
// 	// Get both sequencers ready
// 	actors.ChainA.Sequencer.ActL2PipelineFull(t)
// 	actors.ChainB.Sequencer.ActL2PipelineFull(t)
//
// 	// Set up our user keys
// 	chainAUserKeys := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)
// 	chainBUserKeys := devkeys.ChainUserKeys(actors.ChainB.ChainCfg.ChainID)
// 	userAChainAKey := chainAUserKeys(0)
// 	userAChainBKey := chainBUserKeys(0)
//
// 	// Deploy Emit contract on Chain A
// 	emitContract, emitAddr := deployEmitContract(t, is, actors.ChainA, userAChainAKey)
//
// 	// Set up auth for emitting message
// 	userSecret, err := is.Keys.Secret(userAChainAKey)
// 	require.NoError(t, err)
// 	userAuthA, err := bind.NewKeyedTransactorWithChainID(userSecret, actors.ChainA.ChainCfg.ChainID)
// 	require.NoError(t, err)
//
// 	// Emit test message
// 	testData := []byte("hello chain B!")
// 	tx, err := emitContract.EmitData(userAuthA, testData)
// 	require.NoError(t, err)
//
// 	// Process the emit transaction
// 	actors.ChainA.Sequencer.ActL2StartBlock(t)
// 	err = actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
// 	require.NoError(t, err)
// 	actors.ChainA.Sequencer.ActL2EndBlock(t)
//
// 	// Submit batch and progress L1
// 	actors.ChainA.Batcher.ActSubmitAll(t)
// 	actors.L1Miner.ActL1StartBlock(12)(t)
// 	actors.L1Miner.ActL1IncludeTx(actors.ChainA.BatcherAddr)(t)
// 	actors.L1Miner.ActL1EndBlock(t)
//
// 	// Sync Chain A through the supervisor
// 	actors.Supervisor.SyncEvents(t, actors.ChainA.ChainID)
// 	actors.ChainA.Sequencer.ActL2PipelineFull(t)
// 	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainA.ChainID)
// 	actors.Supervisor.SyncCrossSafe(t, actors.ChainA.ChainID)
//
// 	// Execute message on Chain B
// 	execMessageOnChainB(t, is, actors, emitAddr, testData, userAChainBKey)
// }

// deployEmitContract deploys the Emit contract on the specified chain using the provided key.
func deployEmitContract(t helpers.Testing, is *InteropSetup, chain *Chain, key devkeys.ChainUserKey) (*emit.Emit, common.Address) {
	secret, err := is.Keys.Secret(key)
	require.NoError(t, err)

	deployerAuth, err := bind.NewKeyedTransactorWithChainID(
		secret,
		chain.ChainCfg.ChainID,
	)
	require.NoError(t, err)
	deployerAuth.GasLimit = 3000000
	deployerAuth.GasTipCap = big.NewInt(2 * params.GWei)

	emitAddr, tx, emitContract, err := emit.DeployEmit(deployerAuth, chain.SequencerEngine.EthClient())
	require.NoError(t, err)

	chain.Sequencer.ActL2StartBlock(t)
	err = chain.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(secret.PublicKey))
	require.NoError(t, err)
	chain.Sequencer.ActL2EndBlock(t)

	return emitContract, emitAddr
}

func TestInteropMessageExecution(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	is := SetupInterop(t)
	actors := is.CreateActors()

	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.ChainB.Sequencer.ActL2PipelineFull(t)

	chainAUserKeys := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)
	chainBUserKeys := devkeys.ChainUserKeys(actors.ChainB.ChainCfg.ChainID)
	userAChainAKey := chainAUserKeys(0)
	userAChainBKey := chainBUserKeys(0)

	userSecret, err := is.Keys.Secret(userAChainAKey)
	require.NoError(t, err)
	userAuthA, err := bind.NewKeyedTransactorWithChainID(userSecret, actors.ChainA.ChainCfg.ChainID)
	require.NoError(t, err)

	// Deploy emitter contract and emit a message
	emitContract, emitAddr := deployEmitContract(t, is, actors.ChainA, userAChainAKey)
	testData := []byte("hello chain B!")
	tx, err := emitContract.EmitData(userAuthA, testData)
	require.NoError(t, err)
	t.Logf("Emit transaction hash: %s", tx.Hash().Hex())

	actors.ChainA.Sequencer.ActL2StartBlock(t)
	err = actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
	require.NoError(t, err)
	actors.ChainA.Sequencer.ActL2EndBlock(t)

	receipt, err := actors.ChainA.SequencerEngine.EthClient().TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "emit transaction failed")

	t.Logf("Emit transaction included in block %d at index %d", receipt.BlockNumber.Uint64(), receipt.TransactionIndex)

	var emitEvent *emit.EmitDataEmitted
	for _, log := range receipt.Logs {
		t.Logf("Log: address=%s topics=%v data=%x", log.Address.Hex(), log.Topics, log.Data)
		if log.Address == emitAddr {
			event, err := emitContract.ParseDataEmitted(*log)
			require.NoError(t, err)
			emitEvent = event
			t.Logf("Found DataEmitted event: data=%x", event.Data)
			break
		}
	}
	require.NotNil(t, emitEvent, "DataEmitted event not found")

	actors.ChainA.Batcher.ActSubmitAll(t)
	actors.L1Miner.ActL1StartBlock(12)(t)
	actors.L1Miner.ActL1IncludeTx(actors.ChainA.BatcherAddr)(t)
	actors.L1Miner.ActL1EndBlock(t)

	actors.Supervisor.SyncEvents(t, actors.ChainA.ChainID)
	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainA.ChainID)
	actors.Supervisor.SyncCrossSafe(t, actors.ChainA.ChainID)

	// Log the chain status before execution
	statusA := actors.ChainA.Sequencer.SyncStatus()
	statusB := actors.ChainB.Sequencer.SyncStatus()
	t.Logf("Chain A status - Head: %d Safe: %d Finalized: %d",
		statusA.UnsafeL2.Number, statusA.SafeL2.Number, statusA.FinalizedL2.Number)
	t.Logf("Chain B status - Head: %d Safe: %d Finalized: %d",
		statusB.UnsafeL2.Number, statusB.SafeL2.Number, statusB.FinalizedL2.Number)

	// Setup dependencies
	id, err := actors.ChainA.ChainID.ToUInt32()
	require.NoError(t, err)
	idStr := fmt.Sprintf("%d", id)
	t.Logf("System config proxy address: %s", is.Deployment.L2s[idStr].SystemConfigProxy.Hex())
	contract, err := systemconfig.NewSystemconfig(is.Deployment.L2s[idStr].SystemConfigProxy, actors.ChainB.SequencerEngine.EthClient())
	require.NoError(t, err)
	tx, err = contract.AddDependency(userAuthA, actors.ChainA.ChainCfg.ChainID)
	require.NoError(t, err)

	actors.ChainB.Sequencer.ActL2StartBlock(t)
	err = actors.ChainB.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
	require.NoError(t, err)
	actors.ChainB.Sequencer.ActL2EndBlock(t)

	receipt, err = actors.ChainB.SequencerEngine.EthClient().TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "add dependency transaction failed")

	// Execute the message on Chain B
	execTx := execMessageOnChainB(t, is, actors, emitAddr, receipt.BlockNumber.Uint64(), uint32(receipt.TransactionIndex), testData, userAChainBKey)

	execReceipt, err := actors.ChainB.SequencerEngine.EthClient().TransactionReceipt(context.Background(), execTx.Hash())
	require.NoError(t, err)

	// Log transaction details and any error messages that might be present
	for _, log := range execReceipt.Logs {
		t.Logf("Execution Log: address=%s topics=%v data=%x", log.Address.Hex(), log.Topics, log.Data)
	}

	require.Equal(t, types.ReceiptStatusSuccessful, execReceipt.Status, "execution transaction failed")
}

func execMessageOnChainB(t helpers.Testing, is *InteropSetup, actors *InteropActors,
	emitAddr common.Address, blockNum uint64, logIdx uint32, testData []byte, userKey devkeys.ChainUserKey) *types.Transaction {

	userSecret, err := is.Keys.Secret(userKey)
	require.NoError(t, err)
	userAuth, err := bind.NewKeyedTransactorWithChainID(
		userSecret,
		actors.ChainB.ChainCfg.ChainID,
	)
	require.NoError(t, err)
	userAuth.GasLimit = 3000000
	userAuth.GasTipCap = big.NewInt(2 * params.GWei)

	inboxContract, err := inbox.NewInbox(predeploys.CrossL2InboxAddr, actors.ChainB.SequencerEngine.EthClient())
	require.NoError(t, err)

	status := actors.ChainB.Sequencer.SyncStatus()
	identifier := inbox.Identifier{
		Origin:      emitAddr,
		BlockNumber: new(big.Int).SetUint64(blockNum),
		LogIndex:    new(big.Int).SetUint64(uint64(logIdx)),
		Timestamp:   big.NewInt(int64(status.UnsafeL2.Time)),
		ChainId:     actors.ChainA.ChainCfg.ChainID,
	}

	tx, err := inboxContract.ExecuteMessage(userAuth, identifier, emitAddr, testData)
	require.NoError(t, err)

	actors.ChainB.Sequencer.ActL2StartBlock(t)
	err = actors.ChainB.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
	require.NoError(t, err)
	actors.ChainB.Sequencer.ActL2EndBlock(t)

	actors.Supervisor.SyncEvents(t, actors.ChainB.ChainID)
	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainB.ChainID)
	actors.Supervisor.SyncCrossSafe(t, actors.ChainB.ChainID)

	return tx
}

// - Setup dependecies
// - Fix execution payload

// - Check that executing a message that doesn't exist fails
// - Make sure it doesn't become cross-unsafe

// - Later on reorg
// -
