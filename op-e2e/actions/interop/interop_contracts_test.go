package interop

import (
	"context"
	"crypto/ecdsa"
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

	stypes "github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

// TODO:
//   - Setup dependencies
//   - Check that executing a message that doesn't exist fails
//     - Make sure it doesn't become cross-unsafe
// - Later
//   - Reorg tests

func TestInteropMessageExecutionSuccess(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	is := SetupInterop(t)
	actors := is.CreateActors()

	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.ChainB.Sequencer.ActL2PipelineFull(t)

	// Setup chain B to accept messages from chain A
	setupChainDependency(t, is, actors, actors.ChainA.ChainID)

	// Setup keys and deploy contract
	userKeyA := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)(0)
	userKeyB := devkeys.ChainUserKeys(actors.ChainB.ChainCfg.ChainID)(0)
	secretA, err := is.Keys.Secret(userKeyA)
	require.NoError(t, err)
	secretB, err := is.Keys.Secret(userKeyB)
	require.NoError(t, err)

	_, emitAddr := deployEmitContract(t, is, actors.ChainA, userKeyA)

	// Test successful message passing
	emitResult, execResult := emitAndExecuteMessage(
		t, actors,
		actors.ChainA, actors.ChainB,
		emitAddr, []byte("hello chain B!"),
		secretA, secretB,
	)

	t.Logf("Message emitted in tx %s and executed in tx %s",
		emitResult.Tx.Hash().Hex(), execResult.Tx.Hash().Hex())
}

// TestInteropMessageExecutionFailureNoInitLog tests that attempting to execute a message that doesn't exist fails.
func TestInteropMessageExecutionFailureNoInitLog(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	is := SetupInterop(t)
	actors := is.CreateActors()

	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.ChainB.Sequencer.ActL2PipelineFull(t)

	// Setup chain B to accept messages from chain A
	setupChainDependency(t, is, actors, actors.ChainA.ChainID)

	// Deploy the Emit contract on Chain A to use as a message source
	chainAUserKeys := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)
	chainBUserKeys := devkeys.ChainUserKeys(actors.ChainB.ChainCfg.ChainID)
	userAChainAKey := chainAUserKeys(0)
	userAChainBKey := chainBUserKeys(0)

	_, emitAddr := deployEmitContract(t, is, actors.ChainA, userAChainAKey)

	// Get the user's key for Chain B where we'll attempt the execution
	userAChainBSecret, err := is.Keys.Secret(userAChainBKey)
	require.NoError(t, err)

	// Attempt to execute a non-existent message
	userAuth, err := bind.NewKeyedTransactorWithChainID(userAChainBSecret, actors.ChainB.ChainCfg.ChainID)
	require.NoError(t, err)
	userAuth.GasLimit = 3000000

	// Create message parameters that point to a non-existent event
	fakeIdentifier := inbox.Identifier{
		Origin:      emitAddr,
		BlockNumber: big.NewInt(1),
		LogIndex:    big.NewInt(9999),
		Timestamp:   big.NewInt(1000),
		ChainId:     actors.ChainA.ChainCfg.ChainID,
	}

	// Attempt to execute the non-existent message
	inboxContract, err := inbox.NewInbox(predeploys.CrossL2InboxAddr, actors.ChainB.SequencerEngine.EthClient())
	require.NoError(t, err)

	fakeMessage := []byte("this message was never emitted")
	tx, err := inboxContract.ExecuteMessage(userAuth, fakeIdentifier, emitAddr, fakeMessage)
	require.NoError(t, err)

	// Submit the transaction
	actors.ChainB.Sequencer.ActL2StartBlock(t)
	err = actors.ChainB.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userAChainBSecret.PublicKey))
	require.NoError(t, err)
	actors.ChainB.Sequencer.ActL2EndBlock(t)

	// The transaction should be included but should fail
	receipt, err := actors.ChainB.SequencerEngine.EthClient().TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusFailed, receipt.Status, "execution of non-existent message should fail")
}

// setupChainDependency ensures that the destination chain recognizes the source chain as a dependency
func setupChainDependency(t helpers.Testing, is *InteropSetup, actors *InteropActors, chainID stypes.ChainID) {
	l1ChainID := actors.L1Miner.L1Chain().Config().ChainID

	// Get the system config owner key for the destination chain
	configUserKey := devkeys.ChainOperatorKey{
		ChainID: l1ChainID,
		Role:    devkeys.SystemConfigOwner,
	}
	configUserSecret, err := is.Keys.Secret(configUserKey)
	require.NoError(t, err)

	// Get the system config contract
	id, err := chainID.ToUInt32()
	require.NoError(t, err)
	idStr := fmt.Sprintf("%d", id)
	configContractAddr := is.Deployment.L2s[idStr].SystemConfigProxy
	t.Logf("Setting up dependency for chain %s using SystemConfig at %s",
		idStr, configContractAddr.Hex())

	contract, err := systemconfig.NewSystemconfig(configContractAddr, actors.L1Miner.EthClient())
	require.NoError(t, err)

	// Add the dependency
	auth, err := bind.NewKeyedTransactorWithChainID(configUserSecret, l1ChainID)
	require.NoError(t, err)
	auth.GasLimit = 3000000
	auth.GasTipCap = big.NewInt(2 * params.GWei)

	tx, err := contract.AddDependency(auth, chainID.ToBig())
	require.NoError(t, err)

	// Include the transaction
	actors.L1Miner.ActL1StartBlock(12)(t)
	actors.L1Miner.ActL1IncludeTx(crypto.PubkeyToAddress(configUserSecret.PublicKey))(t)
	actors.L1Miner.ActL1EndBlock(t)

	receipt, err := actors.L1Miner.EthClient().TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status,
		"failed to add chain dependency")
}

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

func serializeLog(log *types.Log) []byte {
	msgPayload := make([]byte, 0)
	for _, topic := range log.Topics {
		msgPayload = append(msgPayload, topic.Bytes()...)
	}
	msgPayload = append(msgPayload, log.Data...)
	return msgPayload
}

type txResult struct {
	Tx      *types.Transaction
	Receipt *types.Receipt
	Logs    []*types.Log
}

func submitL2Transaction(
	t helpers.Testing,
	chain *Chain,
	supervisor *SupervisorActor,
	sender *ecdsa.PrivateKey,
	createTx func(auth *bind.TransactOpts) (*types.Transaction, error),
) *txResult {
	auth, err := bind.NewKeyedTransactorWithChainID(sender, chain.ChainCfg.ChainID)
	require.NoError(t, err)
	auth.GasLimit = 3000000
	auth.GasTipCap = big.NewInt(2 * params.GWei)

	tx, err := createTx(auth)
	require.NoError(t, err)
	t.Logf("Transaction created with hash: %s", tx.Hash().Hex())

	chain.Sequencer.ActL2StartBlock(t)
	err = chain.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(sender.PublicKey))
	require.NoError(t, err)
	chain.Sequencer.ActL2EndBlock(t)

	receipt, err := chain.SequencerEngine.EthClient().TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, types.ReceiptStatusSuccessful, receipt.Status, "transaction failed")

	return &txResult{
		Tx:      tx,
		Receipt: receipt,
		Logs:    receipt.Logs,
	}
}

func emitAndExecuteMessage(
	t helpers.Testing,
	actors *InteropActors,
	srcChain *Chain,
	destChain *Chain,
	emitAddr common.Address,
	msgData []byte,
	sender *ecdsa.PrivateKey,
	executor *ecdsa.PrivateKey,
) (*txResult, *txResult) {
	// Emit message on source chain
	emitResult := submitL2Transaction(t, srcChain, actors.Supervisor, sender, func(auth *bind.TransactOpts) (*types.Transaction, error) {
		emitter, err := emit.NewEmit(emitAddr, srcChain.SequencerEngine.EthClient())
		require.NoError(t, err)
		return emitter.EmitData(auth, msgData)
	})

	// Submit batch and progress L1
	srcChain.Batcher.ActSubmitAll(t)
	actors.L1Miner.ActL1StartBlock(12)(t)
	actors.L1Miner.ActL1IncludeTx(srcChain.BatcherAddr)(t)
	actors.L1Miner.ActL1EndBlock(t)

	// Sync through supervisor
	actors.Supervisor.SyncEvents(t, srcChain.ChainID)
	srcChain.Sequencer.ActL2PipelineFull(t)
	actors.Supervisor.SyncCrossUnsafe(t, srcChain.ChainID)
	actors.Supervisor.SyncCrossSafe(t, srcChain.ChainID)

	// Execute on destination chain
	execResult := submitL2Transaction(t, destChain, actors.Supervisor, executor, func(auth *bind.TransactOpts) (*types.Transaction, error) {
		inboxContract, err := inbox.NewInbox(predeploys.CrossL2InboxAddr, destChain.SequencerEngine.EthClient())
		require.NoError(t, err)

		status := srcChain.Sequencer.SyncStatus()
		identifier := inbox.Identifier{
			Origin:      emitAddr,
			BlockNumber: new(big.Int).SetUint64(emitResult.Receipt.BlockNumber.Uint64()),
			LogIndex:    new(big.Int).SetUint64(uint64(emitResult.Receipt.TransactionIndex)),
			Timestamp:   big.NewInt(int64(status.UnsafeL2.Time)),
			ChainId:     srcChain.ChainCfg.ChainID,
		}

		return inboxContract.ExecuteMessage(auth, identifier, emitAddr, serializeLog(emitResult.Logs[0]))
	})

	return emitResult, execResult
}
