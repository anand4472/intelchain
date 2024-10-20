package nodeconfig

import (
	"testing"

	"github.com/zennittians/intelchain/crypto/bls"

	"github.com/pkg/errors"
	"github.com/zennittians/intelchain/internal/blsgen"
	shardingconfig "github.com/zennittians/intelchain/internal/configs/sharding"
	"github.com/zennittians/intelchain/multibls"
)

func TestNodeConfigSingleton(t *testing.T) {
	// init 3 configs
	_ = GetShardConfig(2)
	// get the singleton variable
	c := GetShardConfig(Global)
	c.SetBeaconGroupID(GroupIDBeacon)
	d := GetShardConfig(Global)
	g := d.GetBeaconGroupID()
	if g != GroupIDBeacon {
		t.Errorf("GetBeaconGroupID = %v, expected = %v", g, GroupIDBeacon)
	}
}

func TestNodeConfigMultiple(t *testing.T) {
	// init 3 configs
	d := GetShardConfig(1)
	e := GetShardConfig(0)
	f := GetShardConfig(42)

	if f != nil {
		t.Errorf("expecting nil, got: %v", f)
	}

	d.SetShardGroupID("abcd")
	if d.GetShardGroupID() != "abcd" {
		t.Errorf("expecting abcd, got: %v", d.GetShardGroupID())
	}

	e.SetClientGroupID("client")
	if e.GetClientGroupID() != "client" {
		t.Errorf("expecting client, got: %v", d.GetClientGroupID())
	}
}

func TestValidateConsensusKeysForSameShard(t *testing.T) {
	// set localnet config
	networkType := "localnet"
	schedule := shardingconfig.LocalnetSchedule
	netType := NetworkType(networkType)
	SetNetworkType(netType)
	SetShardingSchedule(schedule)

	// import two keys that belong to same shard and test ValidateConsensusKeysForSameShard
	keyPath1 := "../../../.itc/dbae2fd6bd97ca23ca1cadf4510327b9f2dce8c1175ad06d46c3df95ef64deeebef70c54664eb6d51384eac9dd9af40a.key"
	priKey1, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath1, "a")
	pubKey1 := priKey1.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	keyPath2 := "../../../.itc/dbae2fd6bd97ca23ca1cadf4510327b9f2dce8c1175ad06d46c3df95ef64deeebef70c54664eb6d51384eac9dd9af40a.key"
	priKey2, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath2, "a")
	pubKey2 := priKey2.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	keys := multibls.PublicKeys{}
	dummyKey := bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey1)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey1, Bytes: dummyKey})
	dummyKey = bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey2)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey2, Bytes: dummyKey})
	if err := GetDefaultConfig().ValidateConsensusKeysForSameShard(keys, 0); err != nil {
		t.Error("expected", nil, "got", err)
	}
	// add third key in different shard and test ValidateConsensusKeysForSameShard
	keyPath3 := "../../../.itc/dbae2fd6bd97ca23ca1cadf4510327b9f2dce8c1175ad06d46c3df95ef64deeebef70c54664eb6d51384eac9dd9af40a.key"
	priKey3, err := blsgen.LoadBLSKeyWithPassPhrase(keyPath3, "a")
	pubKey3 := priKey3.GetPublicKey()
	if err != nil {
		t.Error(err)
	}
	dummyKey = bls.SerializedPublicKey{}
	dummyKey.FromLibBLSPublicKey(pubKey3)
	keys = append(keys, bls.PublicKeyWrapper{Object: pubKey3, Bytes: dummyKey})
	if err := GetDefaultConfig().ValidateConsensusKeysForSameShard(keys, 0); err == nil {
		e := errors.New("bls keys do not belong to the same shard")
		t.Error("expected", e, "got", nil)
	}
}
