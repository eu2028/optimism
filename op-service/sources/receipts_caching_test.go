package sources

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockReceiptsProvider struct {
	mock.Mock
}

func (m *mockReceiptsProvider) FetchReceipts(ctx context.Context, blockInfo eth.BlockInfo, txHashes []common.Hash) (types.Receipts, error) {
	block := eth.ToBlockID(blockInfo)
	args := m.Called(ctx, block, txHashes)
	return args.Get(0).(types.Receipts), args.Error(1)
}

func (m *mockReceiptsProvider) BatchFetchReceipts(ctx context.Context, blockInfos []eth.BlockInfo, txHashes [][]common.Hash) ([]types.Receipts, error) {
	args := m.Called(ctx, blockInfos, txHashes)
	return args.Get(0).(func(infos []eth.BlockInfo) []types.Receipts)(blockInfos), args.Error(1)
}

func TestCachingReceiptsProvider_Caching(t *testing.T) {
	block, receipts := randomRpcBlockAndReceipts(rand.New(rand.NewSource(69)), 4)
	txHashes := receiptTxHashes(receipts)
	blockid := block.BlockID()
	mrp := new(mockReceiptsProvider)
	rp := NewCachingReceiptsProvider(mrp, nil, 1)
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()

	mrp.On("FetchReceipts", ctx, blockid, txHashes).
		Return(types.Receipts(receipts), error(nil)).
		Once() // receipts should be cached after first fetch

	bInfo, _, _ := block.Info(true, true)
	for i := 0; i < 4; i++ {
		gotRecs, err := rp.FetchReceipts(ctx, bInfo, txHashes)
		require.NoError(t, err)
		for i, gotRec := range gotRecs {
			requireEqualReceipt(t, receipts[i], gotRec)
		}
	}
	mrp.AssertExpectations(t)
}

func TestCachingReceiptsProvider_Concurrency(t *testing.T) {
	block, receipts := randomRpcBlockAndReceipts(rand.New(rand.NewSource(69)), 4)
	txHashes := receiptTxHashes(receipts)
	blockid := block.BlockID()
	mrp := new(mockReceiptsProvider)
	rp := NewCachingReceiptsProvider(mrp, nil, 1)

	mrp.On("FetchReceipts", mock.Anything, blockid, txHashes).
		Return(types.Receipts(receipts), error(nil)).
		Once() // receipts should be cached after first fetch

	runConcurrentFetchingTest(t, rp, 32, receipts, block)

	mrp.AssertExpectations(t)
}

func TestCachingReceiptsProvider_Batches(t *testing.T) {
	batch := requestBatch(5)
	blocks := batch.blocks
	infos := batch.infos
	receipts := batch.receipts
	hashes := batch.hashes

	mrp := new(mockReceiptsProvider)
	rp := NewCachingReceiptsProvider(mrp, nil, 1)

	// on the first fetch, the cache should populate with the first block's receipts
	firstID, firstInfo, firstReceipts, firstTxHashes := blocks[0], infos[0], receipts[0], hashes[0]
	mrp.On("FetchReceipts", mock.Anything, firstID, firstTxHashes).
		Return(types.Receipts(firstReceipts), error(nil)).
		Once() // receipts should be cached after first fetch

	// on the batch fetch, we should see only the uncached blocks fetched
	remainingInfos, remainingReceipts, remainingHashes := infos[1:], receipts[1:], hashes[1:]
	mrp.On("BatchFetchReceipts", mock.Anything, remainingInfos, remainingHashes).
		Return(remainingReceipts, error(nil)).
		Once() // receipts should be cached after first fetch

	// fetch
	firstRetReceipts, err := rp.FetchReceipts(context.Background(), firstInfo, firstTxHashes)
	require.NoError(t, err)
	require.Equal(t, firstReceipts, firstRetReceipts)

	// batch fetch
	retReceipts, err := rp.BatchFetchReceipts(context.Background(), infos, hashes)
	require.NoError(t, err)
	// all receipts should be returned in order
	require.Equal(t, receipts, retReceipts)
	mrp.AssertExpectations(t)
}

func TestCachingReceiptsProvider_BatchesConcurrency(t *testing.T) {
	batch := requestBatch(10)
	mrp := new(mockReceiptsProvider)
	rp := NewCachingReceiptsProvider(mrp, nil, 1)

	filter := func(infos []eth.BlockInfo) []types.Receipts {
		result := []types.Receipts{}
		for i := range infos {
			found := false
			for j := range batch.infos {
				if infos[i] == batch.infos[j] {
					result = append(result, batch.receipts[j])
					found = true
					break
				}
			}
			if !found {
				require.Fail(t, "unexpected block info", "info: %v", infos[i])
			}
		}
		return result
	}

	mrp.On("BatchFetchReceipts",
		mock.Anything,
		mock.Anything,
		mock.Anything).
		Return(filter, error(nil))

	runConcurrentBatchFetchingTest(t, rp, 32, batch)
}
