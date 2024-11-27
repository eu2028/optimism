package interfaces

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	bss "github.com/ethereum-optimism/optimism/op-batcher/batcher"
	"github.com/ethereum-optimism/optimism/op-chain-ops/devkeys"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/geth"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/opnode"
	"github.com/ethereum-optimism/optimism/op-e2e/system/helpers"
	l2os "github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor"
	supervisortypes "github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
)

// SuperSystem is an interface for the system (collection of connected resources)
// it provides a way to get the resources for a network by network ID
// and provides a way to get the list of network IDs
// this is useful for testing multiple network backends,
// for example, interopE2ESystem is the default implementation, but a shim to
// kurtosis or another testing framework could be implemented
type SuperSystem interface {
	// get the supervisor
	Supervisor() *supervisor.SupervisorService
	// get the supervisor client
	SupervisorClient() *sources.SupervisorClient
	// get the batcher for a network
	Batcher(network string) *bss.BatcherService
	// get the proposer for a network
	Proposer(network string) *l2os.ProposerService
	// get the opnode for a network
	OpNode(network string) *opnode.Opnode
	// get the geth instance for a network
	L2Geth(network string) *geth.GethInstance
	// get the L2 geth client for a network
	L2GethClient(network string) *ethclient.Client
	// get the secret for a network and role
	L2OperatorKey(network string, role devkeys.ChainOperatorRole) ecdsa.PrivateKey
	// get the list of network IDs as key-strings
	L2IDs() []string
	// get the chain ID for a network
	ChainID(network string) *big.Int
	// register a username to an account on all L2s
	AddUser(username string)
	// get the user key for a user on an L2
	UserKey(id, username string) ecdsa.PrivateKey
	// send a transaction on an L2 on the given network, from the given user
	SendL2Tx(network string, username string, applyTxOpts helpers.TxOptsFn) *types.Receipt
	// get the address for a user on an L2
	Address(network string, username string) common.Address
	// Deploy the Emitter Contract, which emits Event Logs
	DeployEmitterContract(network string, username string) common.Address
	// Use the Emitter Contract to emit an Event Log
	EmitData(network string, username string, data string) *types.Receipt
	// AddDependency adds a dependency (by chain ID) to the given chain
	AddDependency(network string, dep *big.Int) *types.Receipt
	// ExecuteMessage calls the CrossL2Inbox executeMessage function
	ExecuteMessage(
		ctx context.Context,
		id string,
		sender string,
		msgIdentifier supervisortypes.Identifier,
		target common.Address,
		message []byte,
		expectedError error,
	) (*types.Receipt, error)
	// Access a contract on a network by name
	Contract(network string, contractName string) interface{}
}
