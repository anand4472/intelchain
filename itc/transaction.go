package itc

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zennittians/intelchain/core"
	"github.com/zennittians/intelchain/core/rawdb"
	"github.com/zennittians/intelchain/core/types"
	"github.com/zennittians/intelchain/eth/rpc"
)

// SendTx ...
func (itc *Intelchain) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	tx, _, _, _ := rawdb.ReadTransaction(itc.chainDb, signedTx.Hash())
	if tx == nil {
		return itc.NodeAPI.AddPendingTransaction(signedTx)
	}
	return ErrFinalizedTransaction
}

// ResendCx retrieve blockHash from txID and add blockHash to CxPool for resending
// Note that cross shard txn is only for regular txns, not for staking txns, so the input txn hash
// is expected to be regular txn hash
func (itc *Intelchain) ResendCx(ctx context.Context, txID common.Hash) (uint64, bool) {
	blockHash, blockNum, index := itc.BlockChain.ReadTxLookupEntry(txID)
	if blockHash == (common.Hash{}) {
		return 0, false
	}

	blk := itc.BlockChain.GetBlockByHash(blockHash)
	if blk == nil {
		return 0, false
	}

	txs := blk.Transactions()
	// a valid index is from 0 to len-1
	if int(index) > len(txs)-1 {
		return 0, false
	}
	tx := txs[int(index)]

	// check whether it is a valid cross shard tx
	if tx.ShardID() == tx.ToShardID() || blk.Header().ShardID() != tx.ShardID() {
		return 0, false
	}
	entry := core.CxEntry{BlockHash: blockHash, ToShardID: tx.ToShardID()}
	success := itc.CxPool.Add(entry)
	return blockNum, success
}

// GetReceipts ...
func (itc *Intelchain) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return itc.BlockChain.GetReceiptsByHash(hash), nil
}

// GetTransactionsHistory returns list of transactions hashes of address.
func (itc *Intelchain) GetTransactionsHistory(address, txType, order string) ([]common.Hash, error) {
	return itc.NodeAPI.GetTransactionsHistory(address, txType, order)
}

// GetAccountNonce returns the nonce value of the given address for the given block number
func (itc *Intelchain) GetAccountNonce(
	ctx context.Context, address common.Address, blockNum rpc.BlockNumber) (uint64, error) {
	state, _, err := itc.StateAndHeaderByNumber(ctx, blockNum)
	if state == nil || err != nil {
		return 0, err
	}
	return state.GetNonce(address), state.Error()
}

// GetTransactionsCount returns the number of regular transactions of address.
func (itc *Intelchain) GetTransactionsCount(address, txType string) (uint64, error) {
	return itc.NodeAPI.GetTransactionsCount(address, txType)
}

// GetCurrentTransactionErrorSink ..
func (itc *Intelchain) GetCurrentTransactionErrorSink() types.TransactionErrorReports {
	return itc.NodeAPI.ReportPlainErrorSink()
}
