package node

import (
	"errors"
	"testing"

	"github.com/multiformats/go-multiaddr"
	"github.com/stretchr/testify/assert"
	"github.com/zennittians/intelchain/consensus"
	"github.com/zennittians/intelchain/consensus/quorum"
	"github.com/zennittians/intelchain/core"
	"github.com/zennittians/intelchain/crypto/bls"
	"github.com/zennittians/intelchain/internal/chain"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/internal/registry"
	"github.com/zennittians/intelchain/internal/shardchain"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/multibls"
	"github.com/zennittians/intelchain/p2p"
	"github.com/zennittians/intelchain/shard"
)

var testDBFactory = &shardchain.MemDBFactory{}

func TestNewNode(t *testing.T) {
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
	engine := chain.NewEngine()
	decider := quorum.NewDecider(
		quorum.SuperMajorityVote, shard.BeaconChainShardID,
	)
	chainconfig := nodeconfig.GetShardConfig(shard.BeaconChainShardID).GetNetworkType().ChainConfig()
	collection := shardchain.NewCollection(
		nil, testDBFactory, &core.GenesisInitializer{NetworkType: nodeconfig.GetShardConfig(shard.BeaconChainShardID).GetNetworkType()}, engine, &chainconfig,
	)
	blockchain, err := collection.ShardChain(shard.BeaconChainShardID)
	if err != nil {
		t.Fatal("cannot get blockchain")
	}
	reg := registry.New().
		SetBlockchain(blockchain).
		SetEngine(engine).
		SetShardChainCollection(collection)

	consensus, err := consensus.New(
		host, shard.BeaconChainShardID, multibls.GetPrivateKeys(blsKey), reg, decider, 3, false,
	)
	if err != nil {
		t.Fatalf("Cannot craeate consensus: %v", err)
	}

	node := New(host, consensus, nil, nil, nil, nil, reg)
	if node.Consensus == nil {
		t.Error("Consensus is not initialized for the node")
	}

	if node.Blockchain() == nil {
		t.Error("Blockchain is not initialized for the node")
	}

	if node.Blockchain().CurrentBlock() == nil {
		t.Error("Genesis block is not initialized for the node")
	}
}

func TestDNSSyncingPeerProvider(t *testing.T) {
	t.Run("Happy", func(t *testing.T) {
		addrs := make([]multiaddr.Multiaddr, 0)
		p := NewDNSSyncingPeerProvider("example.com", "1234", addrs)
		lookupCount := 0
		lookupName := ""
		p.lookupHost = func(name string) (addrs []string, err error) {
			lookupCount++
			lookupName = name
			return []string{"1.2.3.4", "5.6.7.8"}, nil
		}
		expectedPeers := []p2p.Peer{
			{IP: "1.2.3.4", Port: "1234"},
			{IP: "5.6.7.8", Port: "1234"},
		}
		actualPeers, err := p.SyncingPeers( /*shardID*/ 3)
		if assert.NoError(t, err) {
			assert.Equal(t, actualPeers, expectedPeers)
		}
		assert.Equal(t, lookupCount, 1)
		assert.Equal(t, lookupName, "s3.example.com")
		if err != nil {
			t.Fatalf("SyncingPeers returned non-nil error %#v", err)
		}
	})
	t.Run("LookupError", func(t *testing.T) {
		addrs := make([]multiaddr.Multiaddr, 0)
		p := NewDNSSyncingPeerProvider("example.com", "1234", addrs)
		p.lookupHost = func(_ string) ([]string, error) {
			return nil, errors.New("omg")
		}
		_, actualErr := p.SyncingPeers( /*shardID*/ 3)
		assert.Error(t, actualErr)
	})
}

func TestLocalSyncingPeerProvider(t *testing.T) {
	t.Run("BeaconChain", func(t *testing.T) {
		p := makeLocalSyncingPeerProvider()
		expectedBeaconPeers := []p2p.Peer{
			{IP: "127.0.0.1", Port: "6000"},
			{IP: "127.0.0.1", Port: "6002"},
			{IP: "127.0.0.1", Port: "6004"},
		}
		if actualPeers, err := p.SyncingPeers(0); assert.NoError(t, err) {
			assert.ElementsMatch(t, actualPeers, expectedBeaconPeers)
		}
	})
	t.Run("Shard1Chain", func(t *testing.T) {
		p := makeLocalSyncingPeerProvider()
		expectedShard1Peers := []p2p.Peer{
			// port 6001 omitted because self
			{IP: "127.0.0.1", Port: "6003"},
			{IP: "127.0.0.1", Port: "6005"},
		}
		if actualPeers, err := p.SyncingPeers(1); assert.NoError(t, err) {
			assert.ElementsMatch(t, actualPeers, expectedShard1Peers)
		}
	})
	t.Run("InvalidShard", func(t *testing.T) {
		p := makeLocalSyncingPeerProvider()
		_, err := p.SyncingPeers(999)
		assert.Error(t, err)
	})
}

func makeLocalSyncingPeerProvider() *LocalSyncingPeerProvider {
	return NewLocalSyncingPeerProvider(6000, 6001, 2, 3)
}
