package shardingconfig

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/intelchain-itc/intelchain/crypto/bls"
	"github.com/intelchain-itc/intelchain/internal/genesis"
	"github.com/intelchain-itc/intelchain/numeric"
	"github.com/pkg/errors"
)

// NetworkID is the network type of the blockchain.
type NetworkID byte

// Constants for NetworkID.
const (
	MainNet NetworkID = iota
	TestNet
	LocalNet
	Pangaea
	Partner
	StressNet
	DevNet
)

type instance struct {
	numShards                          uint32
	numNodesPerShard                   int
	numIntelchainOperatedNodesPerShard int
	intelchainVotePercent              numeric.Dec
	externalVotePercent                numeric.Dec
	itcAccounts                        []genesis.DeployAccount
	fnAccounts                         []genesis.DeployAccount
	reshardingEpoch                    []*big.Int
	blocksPerEpoch                     uint64
	slotsLimit                         int // HIP-16: The absolute number of maximum effective slots per shard limit for each validator. 0 means no limit.
	allowlist                          Allowlist
	feeCollectors                      FeeCollectors
	emissionFraction                   numeric.Dec
	recoveryAddress                    ethCommon.Address
}

type FeeCollectors map[ethCommon.Address]numeric.Dec

// NewInstance creates and validates a new sharding configuration based
// upon given parameters.
func NewInstance(
	numShards uint32,
	numNodesPerShard,
	numIntelchainOperatedNodesPerShard,
	slotsLimit int,
	intelchainVotePercent numeric.Dec,
	itcAccounts []genesis.DeployAccount,
	fnAccounts []genesis.DeployAccount,
	allowlist Allowlist,
	feeCollectors FeeCollectors,
	emissionFractionToRecovery numeric.Dec,
	recoveryAddress ethCommon.Address,
	reshardingEpoch []*big.Int, blocksE uint64,
) (Instance, error) {
	if numShards < 1 {
		return nil, errors.Errorf(
			"sharding config must have at least one shard have %d", numShards,
		)
	}
	if numNodesPerShard < 1 {
		return nil, errors.Errorf(
			"each shard must have at least one node %d", numNodesPerShard,
		)
	}
	if numIntelchainOperatedNodesPerShard < 0 {
		return nil, errors.Errorf(
			"Intelchain-operated nodes cannot be negative %d", numIntelchainOperatedNodesPerShard,
		)
	}
	if numIntelchainOperatedNodesPerShard > numNodesPerShard {
		return nil, errors.Errorf(""+
			"number of Intelchain-operated nodes cannot exceed "+
			"overall number of nodes per shard %d %d",
			numIntelchainOperatedNodesPerShard,
			numNodesPerShard,
		)
	}
	if slotsLimit < 0 {
		return nil, errors.Errorf("SlotsLimit cannot be negative %d", slotsLimit)
	}
	if intelchainVotePercent.LT(numeric.ZeroDec()) ||
		intelchainVotePercent.GT(numeric.OneDec()) {
		return nil, errors.Errorf("" +
			"total voting power of Intelchain nodes should be within [0, 1]",
		)
	}
	if len(feeCollectors) > 0 {
		total := numeric.ZeroDec() // is a copy
		for _, v := range feeCollectors {
			total = total.Add(v)
		}
		if !total.Equal(numeric.OneDec()) {
			return nil, errors.Errorf(
				"total fee collection percentage should be 1, but got %v", total,
			)
		}
	}
	if emissionFractionToRecovery.LT(numeric.ZeroDec()) ||
		emissionFractionToRecovery.GT(numeric.OneDec()) {
		return nil, errors.Errorf(
			"emission split must be within [0, 1]",
		)
	}
	if !emissionFractionToRecovery.Equal(numeric.ZeroDec()) {
		if recoveryAddress == (ethCommon.Address{}) {
			return nil, errors.Errorf(
				"have non-zero emission split but no target address",
			)
		}
	}
	if recoveryAddress != (ethCommon.Address{}) {
		if emissionFractionToRecovery.Equal(numeric.ZeroDec()) {
			return nil, errors.Errorf(
				"have target address but no emission split",
			)
		}
	}

	return instance{
		numShards:                          numShards,
		numNodesPerShard:                   numNodesPerShard,
		numIntelchainOperatedNodesPerShard: numIntelchainOperatedNodesPerShard,
		intelchainVotePercent:              intelchainVotePercent,
		externalVotePercent:                numeric.OneDec().Sub(intelchainVotePercent),
		itcAccounts:                        itcAccounts,
		fnAccounts:                         fnAccounts,
		allowlist:                          allowlist,
		reshardingEpoch:                    reshardingEpoch,
		blocksPerEpoch:                     blocksE,
		slotsLimit:                         slotsLimit,
		feeCollectors:                      feeCollectors,
		recoveryAddress:                    recoveryAddress,
		emissionFraction:                   emissionFractionToRecovery,
	}, nil
}

// MustNewInstance creates a new sharding configuration based upon
// given parameters.  It panics if parameter validation fails.
// It is intended to be used for static initialization.
func MustNewInstance(
	numShards uint32,
	numNodesPerShard, numIntelchainOperatedNodesPerShard int, slotsLimitPercent float32,
	intelchainVotePercent numeric.Dec,
	itcAccounts []genesis.DeployAccount,
	fnAccounts []genesis.DeployAccount,
	allowlist Allowlist,
	feeCollectors FeeCollectors,
	emissionFractionToRecovery numeric.Dec,
	recoveryAddress ethCommon.Address,
	reshardingEpoch []*big.Int, blocksPerEpoch uint64,
) Instance {
	slotsLimit := int(float32(numNodesPerShard-numIntelchainOperatedNodesPerShard) * slotsLimitPercent)
	sc, err := NewInstance(
		numShards, numNodesPerShard, numIntelchainOperatedNodesPerShard,
		slotsLimit, intelchainVotePercent, itcAccounts, fnAccounts,
		allowlist, feeCollectors, emissionFractionToRecovery,
		recoveryAddress, reshardingEpoch, blocksPerEpoch,
	)
	if err != nil {
		panic(err)
	}
	return sc
}

// BlocksPerEpoch ..
func (sc instance) BlocksPerEpoch() uint64 {
	return sc.blocksPerEpoch
}

// NumShards returns the number of shards in the network.
func (sc instance) NumShards() uint32 {
	return sc.numShards
}

// SlotsLimit returns the max slots per shard limit for each validator
func (sc instance) SlotsLimit() int {
	return sc.slotsLimit
}

// FeeCollector returns a mapping of address to decimal % of fee
func (sc instance) FeeCollectors() FeeCollectors {
	return sc.feeCollectors
}

// IntelchainVotePercent returns total percentage of voting power Intelchain nodes possess.
func (sc instance) IntelchainVotePercent() numeric.Dec {
	return sc.intelchainVotePercent
}

// ExternalVotePercent returns total percentage of voting power external validators possess.
func (sc instance) ExternalVotePercent() numeric.Dec {
	return sc.externalVotePercent
}

// NumNodesPerShard returns number of nodes in each shard.
func (sc instance) NumNodesPerShard() int {
	return sc.numNodesPerShard
}

// NumIntelchainOperatedNodesPerShard returns number of nodes in each shard
// that are operated by Intelchain.
func (sc instance) NumIntelchainOperatedNodesPerShard() int {
	return sc.numIntelchainOperatedNodesPerShard
}

// ItcAccounts returns the list of Intelchain accounts
func (sc instance) ItcAccounts() []genesis.DeployAccount {
	return sc.itcAccounts
}

// FnAccounts returns the list of Foundational Node accounts
func (sc instance) FnAccounts() []genesis.DeployAccount {
	return sc.fnAccounts
}

// FindAccount returns the deploy account based on the blskey, and if the account is a leader
// or not in the bootstrapping process.
func (sc instance) FindAccount(blsPubKey string) (bool, *genesis.DeployAccount) {
	for i, item := range sc.itcAccounts {
		if item.BLSPublicKey == blsPubKey {
			item.ShardID = uint32(i) % sc.numShards
			return uint32(i) < sc.numShards, &item
		}
	}
	for i, item := range sc.fnAccounts {
		if item.BLSPublicKey == blsPubKey {
			item.ShardID = uint32(i) % sc.numShards
			return false, &item
		}
	}
	return false, nil
}

// ReshardingEpoch returns the list of epoch number
func (sc instance) ReshardingEpoch() []*big.Int {
	return sc.reshardingEpoch
}

// ReshardingEpoch returns the list of epoch number
func (sc instance) GetNetworkID() NetworkID {
	return DevNet
}

// ExternalAllowlist returns the list of external leader keys in allowlist(HIP18)
func (sc instance) ExternalAllowlist() []bls.PublicKeyWrapper {
	return sc.allowlist.BLSPublicKeys
}

// ExternalAllowlistLimit returns the maximum number of external leader keys on each shard
func (sc instance) ExternalAllowlistLimit() int {
	return sc.allowlist.MaxLimitPerShard
}

func (sc instance) HIP30RecoveryAddress() ethCommon.Address {
	return sc.recoveryAddress
}

func (sc instance) HIP30EmissionFraction() numeric.Dec {
	return sc.emissionFraction
}
