package worker

import (
	"math/big"
	"math/rand"
	"testing"

	"github.com/intelchain-itc/intelchain/core/rawdb"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	blockfactory "github.com/intelchain-itc/intelchain/block/factory"
	"github.com/intelchain-itc/intelchain/common/denominations"
	"github.com/intelchain-itc/intelchain/core"
	"github.com/intelchain-itc/intelchain/core/types"
	"github.com/intelchain-itc/intelchain/core/vm"
	chain2 "github.com/intelchain-itc/intelchain/internal/chain"
	"github.com/intelchain-itc/intelchain/internal/params"
)

var (
	// Test accounts
	testBankKey, _  = crypto.GenerateKey()
	testBankAddress = crypto.PubkeyToAddress(testBankKey.PublicKey)
	testBankFunds   = big.NewInt(8000000000000000000)

	chainConfig  = params.TestChainConfig
	blockFactory = blockfactory.ForTest
)

func TestNewWorker(t *testing.T) {
	// Setup a new blockchain with genesis block containing test token on test address
	var (
		database = rawdb.NewMemoryDatabase()
		gspec    = core.Genesis{
			Config:  chainConfig,
			Factory: blockFactory,
			Alloc:   core.GenesisAlloc{testBankAddress: {Balance: testBankFunds}},
			ShardID: 10,
		}
		engine = chain2.NewEngine()
	)

	genesis := gspec.MustCommit(database)
	_ = genesis
	cacheConfig := &core.CacheConfig{SnapshotLimit: 0}
	chain, err := core.NewBlockChain(database, nil, &core.BlockChainImpl{}, cacheConfig, gspec.Config, engine, vm.Config{})

	if err != nil {
		t.Error(err)
	}
	// Create a new worker
	worker := New(chain, nil)

	if worker.GetCurrentState().GetBalance(crypto.PubkeyToAddress(testBankKey.PublicKey)).Cmp(testBankFunds) != 0 {
		t.Error("Worker state is not setup correctly")
	}
}

func TestCommitTransactions(t *testing.T) {
	// Setup a new blockchain with genesis block containing test token on test address
	var (
		database = rawdb.NewMemoryDatabase()
		gspec    = core.Genesis{
			Config:  chainConfig,
			Factory: blockFactory,
			Alloc:   core.GenesisAlloc{testBankAddress: {Balance: testBankFunds}},
			ShardID: 0,
		}
		engine = chain2.NewEngine()
	)

	gspec.MustCommit(database)
	cacheConfig := &core.CacheConfig{SnapshotLimit: 0}
	chain, _ := core.NewBlockChain(database, nil, nil, cacheConfig, gspec.Config, engine, vm.Config{})

	// Create a new worker
	worker := New(chain, nil)

	// Generate a test tx
	baseNonce := worker.GetCurrentState().GetNonce(crypto.PubkeyToAddress(testBankKey.PublicKey))
	randAmount := rand.Float32()
	tx, _ := types.SignTx(types.NewTransaction(baseNonce, testBankAddress, uint32(0), big.NewInt(int64(denominations.ITC)*int64(randAmount)), 21000, big.NewInt(1), nil), types.HomesteadSigner{}, testBankKey)

	// Commit the tx to the worker
	txs := make(map[common.Address]types.Transactions)
	txs[testBankAddress] = types.Transactions{tx}
	err := worker.CommitTransactions(
		txs, nil, testBankAddress,
	)
	if err != nil {
		t.Error(err)
	}

	if len(worker.GetCurrentReceipts()) == 0 {
		t.Error("No receipt is created for new transactions")
	}

	if len(worker.current.txs) != 1 {
		t.Error("Transaction is not committed")
	}
}

func TestGasLimit(t *testing.T) {
	w := newWorker(
		&params.ChainConfig{
			BlockGas30MEpoch: big.NewInt(10),
		}, nil, nil)
	require.EqualValues(t, 80_000_000, w.GasFloor(big.NewInt(3)))
	require.EqualValues(t, 30_000_000, w.GasFloor(big.NewInt(10)))
}
