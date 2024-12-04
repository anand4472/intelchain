package node

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/intelchain-itc/intelchain/consensus"
	"github.com/intelchain-itc/intelchain/consensus/quorum"
	"github.com/intelchain-itc/intelchain/core"
	"github.com/intelchain-itc/intelchain/core/types"
	"github.com/intelchain-itc/intelchain/crypto/bls"
	"github.com/intelchain-itc/intelchain/internal/chain"
	nodeconfig "github.com/intelchain-itc/intelchain/internal/configs/node"
	"github.com/intelchain-itc/intelchain/internal/registry"
	"github.com/intelchain-itc/intelchain/internal/shardchain"
	"github.com/intelchain-itc/intelchain/internal/utils"
	"github.com/intelchain-itc/intelchain/multibls"
	"github.com/intelchain-itc/intelchain/p2p"
	"github.com/intelchain-itc/intelchain/shard"
	staking "github.com/intelchain-itc/intelchain/staking/types"
	"github.com/stretchr/testify/require"
)

func TestFinalizeNewBlockAsync(t *testing.T) {
	blsKey := bls.RandPrivateKey()
	pubKey := blsKey.GetPublicKey()
	leader := p2p.Peer{IP: "127.0.0.1", Port: "8882", ConsensusPubKey: pubKey}
	priKey, _, _ := utils.GenKeyP2P("127.0.0.1", "9902")
	host, err := p2p.NewHost(p2p.HostConfig{
		Self:   &leader,
		BLSKey: priKey,
	})
	if err != nil {
		t.Fatalf("newhost failure: %v", err)
	}
	var testDBFactory = &shardchain.MemDBFactory{}
	engine := chain.NewEngine()
	chainconfig := nodeconfig.GetShardConfig(shard.BeaconChainShardID).GetNetworkType().ChainConfig()
	collection := shardchain.NewCollection(
		nil, testDBFactory, &core.GenesisInitializer{NetworkType: nodeconfig.GetShardConfig(shard.BeaconChainShardID).GetNetworkType()}, engine, &chainconfig,
	)
	blockchain, err := collection.ShardChain(shard.BeaconChainShardID)
	require.NoError(t, err)

	decider := quorum.NewDecider(
		quorum.SuperMajorityVote, shard.BeaconChainShardID,
	)
	reg := registry.New().
		SetBlockchain(blockchain).
		SetBeaconchain(blockchain).
		SetEngine(engine).
		SetShardChainCollection(collection)
	consensusObj, err := consensus.New(
		host, shard.BeaconChainShardID, multibls.GetPrivateKeys(blsKey), reg, decider, 3, false,
	)
	if err != nil {
		t.Fatalf("Cannot craeate consensus: %v", err)
	}

	node := New(host, consensusObj, nil, nil, nil, nil, reg)

	node.Worker.UpdateCurrent()

	txs := make(map[common.Address]types.Transactions)
	stks := staking.StakingTransactions{}
	node.Worker.CommitTransactions(
		txs, stks, common.Address{},
	)
	commitSigs := make(chan []byte, 1)
	commitSigs <- []byte{}

	block, _ := node.Worker.FinalizeNewBlock(
		commitSigs, func() uint64 { return 0 }, common.Address{}, nil, nil,
	)

	if err := consensus.VerifyNewBlock(nil, blockchain, nil)(block); err != nil {
		t.Error("New block is not verified successfully:", err)
	}

	node.Blockchain().InsertChain(types.Blocks{block}, false)

	node.Worker.UpdateCurrent()

	_, err = node.Worker.FinalizeNewBlock(
		commitSigs, func() uint64 { return 0 }, common.Address{}, nil, nil,
	)

	if !strings.Contains(err.Error(), "cannot finalize block") {
		t.Error("expect timeout on FinalizeNewBlock")
	}
}
