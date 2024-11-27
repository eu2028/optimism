package interop

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-e2e/actions/helpers"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/interop/contracts/bindings/emit"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/interop/contracts/bindings/inbox"
	"github.com/ethereum-optimism/optimism/op-service/predeploys"
)

func TestInteropMessageExecution(gt *testing.T) {
	t := helpers.NewDefaultTesting(gt)
	is := SetupInterop(t)
	actors := is.CreateActors()

	// Get both sequencers ready
	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.ChainB.Sequencer.ActL2PipelineFull(t)

	// We'll use User 0 for our testing
	secret, err := is.Keys.Secret(devkeys.ChainOperatorKey{
		ChainID: actors.ChainA.ChainCfg.ChainID,
		Role:    devkeys.ProposerRole,
	})

	// Create transaction options for the user
	deployerAuth, err := bind.NewKeyedTransactorWithChainID(
		secret,
		actors.ChainA.ChainCfg.ChainID,
	)
	require.NoError(t, err)
	deployerAuth.GasLimit = 3000000
	deployerAuth.GasTipCap = big.NewInt(2 * params.GWei)

	// Deploy the Emit contract on Chain A
	emitAddr, tx, emitContract, err := emit.DeployEmit(deployerAuth, actors.ChainA.SequencerEngine.EthClient())
	require.NoError(t, err)

	// Include the deploy transaction
	actors.ChainA.Sequencer.ActL2StartBlock(t)
	actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(secret.PublicKey))
	actors.ChainA.Sequencer.ActL2EndBlock(t)

	// Create test user

	userChainKey := devkeys.ChainUserKeys(actors.ChainA.ChainCfg.ChainID)(0)
	userSecret, err := is.Keys.Secret(userChainKey)
	userAuthA, err := bind.NewKeyedTransactorWithChainID(
		userSecret,
		actors.ChainA.ChainCfg.ChainID,
	)
	require.NoError(t, err)

	// Create and emit the test message
	testData := []byte("hello chain B!")
	tx, err = emitContract.EmitData(userAuthA, testData)
	require.NoError(t, err)

	// Include the emit transaction
	actors.ChainA.Sequencer.ActL2StartBlock(t)
	err = actors.ChainA.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(secret.PublicKey))
	require.NoError(t, err)
	actors.ChainA.Sequencer.ActL2EndBlock(t)

	// Submit batch and mine it
	actors.ChainA.Batcher.ActSubmitAll(t)
	actors.L1Miner.ActL1StartBlock(12)(t)
	actors.L1Miner.ActL1IncludeTx(actors.ChainA.BatcherAddr)(t)
	actors.L1Miner.ActL1EndBlock(t)

	// Sync Chain A status to supervisor and wait for safety
	actors.Supervisor.SyncEvents(t, actors.ChainA.ChainID)
	actors.ChainA.Sequencer.ActL2PipelineFull(t)
	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainA.ChainID)
	actors.Supervisor.SyncCrossSafe(t, actors.ChainA.ChainID)

	// Now try to execute the message on Chain B using the same key
	userAuthB, err := bind.NewKeyedTransactorWithChainID(
		userSecret,
		actors.ChainB.ChainCfg.ChainID,
	)
	require.NoError(t, err)
	require.NoError(t, err)
	userAuthB.GasLimit = 3000000
	userAuthB.GasTipCap = big.NewInt(2 * params.GWei)

	inboxContract, err := inbox.NewInbox(predeploys.CrossL2InboxAddr, actors.ChainB.SequencerEngine.EthClient())
	require.NoError(t, err)

	// Get the current status of Chain B
	status := actors.ChainB.Sequencer.SyncStatus()

	// Execute the message
	identifier := inbox.Identifier{
		Origin:      emitAddr,
		BlockNumber: big.NewInt(2), // TODO: Get actual block number
		LogIndex:    big.NewInt(0),
		Timestamp:   big.NewInt(int64(status.UnsafeL2.Time)),
		ChainId:     actors.ChainA.ChainCfg.ChainID,
	}

	tx, err = inboxContract.ExecuteMessage(userAuthB, identifier, emitAddr, testData)
	require.NoError(t, err)

	// Include the execute transaction
	actors.ChainB.Sequencer.ActL2StartBlock(t)
	err = actors.ChainB.SequencerEngine.EngineApi.IncludeTx(tx, crypto.PubkeyToAddress(userSecret.PublicKey))
	require.NoError(t, err)
	actors.ChainB.Sequencer.ActL2EndBlock(t)

	// Sync Chain B and verify message execution
	actors.Supervisor.SyncEvents(t, actors.ChainB.ChainID)
	actors.Supervisor.SyncCrossUnsafe(t, actors.ChainB.ChainID)
	actors.Supervisor.SyncCrossSafe(t, actors.ChainB.ChainID)
}
