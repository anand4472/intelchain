package itc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zennittians/intelchain/core/types"
)

// GetPoolStats returns the number of pending and queued transactions
func (itc *Intelchain) GetPoolStats() (pendingCount, queuedCount int) {
	return itc.TxPool.Stats()
}

// GetPoolNonce ...
func (itc *Intelchain) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return itc.TxPool.State().GetNonce(addr), nil
}

// GetPoolTransaction ...
func (itc *Intelchain) GetPoolTransaction(hash common.Hash) types.PoolTransaction {
	return itc.TxPool.Get(hash)
}

// GetPendingCXReceipts ..
func (itc *Intelchain) GetPendingCXReceipts() []*types.CXReceiptsProof {
	return itc.NodeAPI.PendingCXReceipts()
}

// GetPoolTransactions returns pool transactions.
func (itc *Intelchain) GetPoolTransactions() (types.PoolTransactions, error) {
	pending, err := itc.TxPool.Pending()
	if err != nil {
		return nil, err
	}
	queued, err := itc.TxPool.Queued()
	if err != nil {
		return nil, err
	}
	var txs types.PoolTransactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	for _, batch := range queued {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (itc *Intelchain) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return itc.gpo.SuggestPrice(ctx)
}
