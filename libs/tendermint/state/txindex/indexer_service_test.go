package txindex_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/tendermint/tm-db"

	abci "github.com/okex/exchain/libs/tendermint/abci/types"
	"github.com/okex/exchain/libs/tendermint/libs/log"
	"github.com/okex/exchain/libs/tendermint/state/txindex"
	"github.com/okex/exchain/libs/tendermint/state/txindex/kv"
	"github.com/okex/exchain/libs/tendermint/types"
)

func TestIndexerServiceIndexesBlocks(t *testing.T) {
	// event bus
	eventBus := types.NewEventBus()
	eventBus.SetLogger(log.TestingLogger())
	err := eventBus.Start()
	require.NoError(t, err)
	defer eventBus.Stop()

	// tx indexer
	store := db.NewMemDB()
	txIndexer := kv.NewTxIndex(store, kv.IndexAllEvents())

	service := txindex.NewIndexerService(txIndexer, eventBus)
	service.SetLogger(log.TestingLogger())
	err = service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// publish block with txs
	eventBus.PublishEventNewBlockHeader(types.EventDataNewBlockHeader{
		Header: types.Header{Height: 1},
		NumTxs: int64(2),
	})
	txResult1 := &types.TxResult{
		Height: 1,
		Index:  uint32(0),
		Tx:     types.Tx("foo"),
		Result: abci.ResponseDeliverTx{Code: 0},
	}
	eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult1})
	txResult2 := &types.TxResult{
		Height: 1,
		Index:  uint32(1),
		Tx:     types.Tx("bar"),
		Result: abci.ResponseDeliverTx{Code: 0},
	}
	eventBus.PublishEventTx(types.EventDataTx{TxResult: *txResult2})

	time.Sleep(100 * time.Millisecond)

	// check the result
	res, err := txIndexer.Get(types.Tx("foo").Hash())
	assert.NoError(t, err)
	assert.Equal(t, txResult1, res)
	res, err = txIndexer.Get(types.Tx("bar").Hash())
	assert.NoError(t, err)
	assert.Equal(t, txResult2, res)
}
