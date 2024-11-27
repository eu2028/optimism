package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/interop2/testing/interfaces"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-supervisor/supervisor/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"

	gethTypes "github.com/ethereum/go-ethereum/core/types"
)

type SuperSystemAutomation struct {
	Sys    interfaces.SuperSystem
	Logger log.Logger
	T      interfaces.Test

	users  []string
	chains []string

	rollupClients map[string]*sources.RollupClient
	mtx           sync.RWMutex
}

func NewSuperSystemAutomation(sys interfaces.SuperSystem, logger log.Logger, t interfaces.Test) *SuperSystemAutomation {
	return &SuperSystemAutomation{
		Sys:    sys,
		Logger: logger,
		T:      t,

		chains: sys.L2IDs(),
	}
}

type SyncPoint struct {
	ev    *gethTypes.Log
	chain string
	auto  *SuperSystemAutomation
}

func (sp *SyncPoint) Event() *gethTypes.Log {
	return sp.ev
}

func (sp *SyncPoint) Identifier() types.Identifier {
	ethCl := sp.auto.Sys.L2GethClient(sp.chain)
	header, err := ethCl.HeaderByHash(context.Background(), sp.ev.BlockHash)
	require.NoError(sp.auto.T, err)

	return types.Identifier{
		Origin:      sp.ev.Address,
		BlockNumber: sp.ev.BlockNumber,
		LogIndex:    uint32(sp.ev.Index),
		Timestamp:   header.Time,
		ChainID:     types.ChainIDFromBig(sp.auto.Sys.ChainID(sp.chain)),
	}
}

func (s *SuperSystemAutomation) GetRollupClient(chain string) (*sources.RollupClient, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.rollupClients == nil {
		s.rollupClients = make(map[string]*sources.RollupClient)
	}

	if client, ok := s.rollupClients[chain]; ok {
		return client, nil
	}

	rpc := s.Sys.OpNode(chain).UserRPC().RPC()
	client, err := dial.DialRollupClientWithTimeout(context.Background(), time.Second*15, s.Logger, rpc)
	if err != nil {
		return nil, err
	}
	s.rollupClients[chain] = client
	return client, nil
}

func (s *SuperSystemAutomation) addUser(name string) {
	s.Sys.AddUser(name)
	s.users = append(s.users, name)
}

func nameGenerator(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func (s *SuperSystemAutomation) NewUniqueUser(prefix string) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	name := nameGenerator(prefix)
	s.addUser(name)
	return name
}

func (s *SuperSystemAutomation) NewUniqueUsers(n int) []string {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	names := make([]string, n)
	for i := 0; i < n; i++ {
		name := nameGenerator(fmt.Sprintf("User%d", i))
		names[i] = name
		s.addUser(name)
	}
	return names
}

func (s *SuperSystemAutomation) User(idx int) string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	require.Less(s.T, idx, len(s.users), "user index out of bounds")
	return s.users[idx]
}

func (s *SuperSystemAutomation) Chain(idx int) string {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	require.Less(s.T, idx, len(s.chains), "chain index out of bounds")
	return s.chains[idx]
}

func (s *SuperSystemAutomation) SetupXChainMessaging(sender string, orig string, dest string) error {
	s.Sys.DeployEmitterContract(orig, sender)
	depRec := s.Sys.AddDependency(dest, s.Sys.ChainID(orig))

	rollupClA, err := s.GetRollupClient(orig)
	if err != nil {
		return err
	}

	// Now wait for the dependency to be visible in the L2 (receipt needs to be picked up)
	require.Eventually(s.T, func() bool {
		status, err := rollupClA.SyncStatus(context.Background())
		require.NoError(s.T, err)
		return status.CrossUnsafeL2.L1Origin.Number >= depRec.BlockNumber.Uint64()
	}, time.Second*30, time.Second, "wait for L1 origin to match dependency L1 block")

	return nil
}

func (s *SuperSystemAutomation) SendXChainMessage(sender string, chain string, data string) (*SyncPoint, error) {
	emitRec := s.Sys.EmitData(chain, sender, data)
	s.T.Logf("Emitted a log event in block %d", emitRec.BlockNumber.Uint64())

	// Wait for initiating side to become cross-unsafe
	require.Eventually(s.T, func() bool {
		rollupCl, err := s.GetRollupClient(chain)
		require.NoError(s.T, err)
		status, err := rollupCl.SyncStatus(context.Background())
		require.NoError(s.T, err)
		return status.CrossUnsafeL2.Number >= emitRec.BlockNumber.Uint64()
	}, time.Second*60, time.Second, "wait for emitted data to become cross-unsafe")
	s.T.Logf("Reached cross-unsafe block %d", emitRec.BlockNumber.Uint64())

	require.Len(s.T, emitRec.Logs, 1)
	ev := emitRec.Logs[0]
	return &SyncPoint{ev: ev, chain: chain, auto: s}, nil
}
