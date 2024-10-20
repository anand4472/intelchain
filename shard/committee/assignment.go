package committee

import (
	"encoding/json"
	"math/big"
	"fmt"
	"github.com/zennittians/intelchain/core/state"
	"github.com/zennittians/intelchain/crypto/bls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	bls_core "github.com/zennittians/bls/ffi/go/bls"
	"github.com/zennittians/intelchain/block"
	"github.com/zennittians/intelchain/core/types"
	common2 "github.com/zennittians/intelchain/internal/common"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	shardingconfig "github.com/zennittians/intelchain/internal/configs/sharding"
	"github.com/zennittians/intelchain/internal/params"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/numeric"
	"github.com/zennittians/intelchain/shard"
	"github.com/zennittians/intelchain/staking/availability"
	"github.com/zennittians/intelchain/staking/effective"
	staking "github.com/zennittians/intelchain/staking/types"
)

// StakingCandidatesReader ..
type StakingCandidatesReader interface {
	CurrentBlock() *types.Block
	StateAt(root common.Hash) (*state.DB, error)
	ReadValidatorInformation(addr common.Address) (*staking.ValidatorWrapper, error)
	ReadValidatorInformationAtState(
		addr common.Address, state *state.DB,
	) (*staking.ValidatorWrapper, error)
	ReadValidatorSnapshot(addr common.Address) (*staking.ValidatorSnapshot, error)
	ValidatorCandidates() []common.Address
}

// CandidatesForEPoS ..
type CandidatesForEPoS struct {
	Orders                             map[common.Address]effective.SlotOrder
	OpenSlotCountForExternalValidators int
}

// CompletedEPoSRound ..
type CompletedEPoSRound struct {
	MedianStake         numeric.Dec              `json:"epos-median-stake"`
	MaximumExternalSlot int                      `json:"max-external-slots"`
	AuctionWinners      []effective.SlotPurchase `json:"epos-slot-winners"`
	AuctionCandidates   []*CandidateOrder        `json:"epos-slot-candidates"`
}

// CandidateOrder ..
type CandidateOrder struct {
	*effective.SlotOrder
	StakePerKey *big.Int
	Validator   common.Address
}

// MarshalJSON ..
func (p CandidateOrder) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		*effective.SlotOrder
		StakePerKey *big.Int `json:"stake-per-key"`
		Validator   string   `json:"validator"`
	}{
		p.SlotOrder,
		p.StakePerKey,
		common2.MustAddressToBech32(p.Validator), // Updated to use single argument
	})
}

// NewEPoSRound runs a fresh computation of EPoS using
// latest data always
func NewEPoSRound(epoch *big.Int, stakedReader StakingCandidatesReader, isExtendedBound bool, slotsLimit, shardCount int) (
	*CompletedEPoSRound, error,
) {
	eligibleCandidate, err := prepareOrders(stakedReader, slotsLimit, shardCount)
	if err != nil {
		return nil, err
	}
	maxExternalSlots := shard.ExternalSlotsAvailableForEpoch(
		epoch,
	)
	median, winners := effective.Apply(
		eligibleCandidate, maxExternalSlots, isExtendedBound,
	)
	auctionCandidates := make([]*CandidateOrder, len(eligibleCandidate))

	i := 0
	for key := range eligibleCandidate {
		perKey := big.NewInt(0)
		if l := len(eligibleCandidate[key].SpreadAmong); l > 0 {
			perKey.Set(
				new(big.Int).Div(
					eligibleCandidate[key].Stake, big.NewInt(int64(l)),
				),
			)
		}
		auctionCandidates[i] = &CandidateOrder{
			SlotOrder:   eligibleCandidate[key],
			StakePerKey: perKey,
			Validator:   key,
		}
		i++
	}

	return &CompletedEPoSRound{
		MedianStake:         median,
		MaximumExternalSlot: maxExternalSlots,
		AuctionWinners:      winners,
		AuctionCandidates:   auctionCandidates,
	}, nil
}

func prepareOrders(
	stakedReader StakingCandidatesReader,
	slotsLimit, shardCount int,
) (map[common.Address]*effective.SlotOrder, error) {
	candidates := stakedReader.ValidatorCandidates()
	blsKeys := map[bls.SerializedPublicKey]struct{}{}
	essentials := map[common.Address]*effective.SlotOrder{}
	totalStaked, tempZero := big.NewInt(0), numeric.ZeroDec()

	instance := shard.Schedule.InstanceForEpoch(stakedReader.CurrentBlock().Epoch())
	for _, account := range instance.ItcAccounts() {
		pub := &bls_core.PublicKey{}
		if err := pub.DeserializeHexStr(account.BLSPublicKey); err != nil {
			continue
		}
		pubKey := bls.SerializedPublicKey{}
		if err := pubKey.FromLibBLSPublicKey(pub); err != nil {
			continue
		}
		blsKeys[pubKey] = struct{}{}
	}

	state, err := stakedReader.StateAt(stakedReader.CurrentBlock().Root())
	if err != nil || state == nil {
		return nil, errors.Wrapf(err, "not state found at root: %s", stakedReader.CurrentBlock().Root().Hex())
	}
	for i := range candidates {
		validator, err := stakedReader.ReadValidatorInformationAtState(
			candidates[i], state,
		)
		if err != nil {
			return nil, err
		}
		snapshot, err := stakedReader.ReadValidatorSnapshot(
			candidates[i],
		)
		if err != nil {
			return nil, err
		}
		if !IsEligibleForEPoSAuction(snapshot, validator) {
			continue
		}

		slotPubKeysLimited := make([]bls.SerializedPublicKey, 0, len(validator.SlotPubKeys))
		found := false
		shardSlotsCount := make([]int, shardCount)
		for _, key := range validator.SlotPubKeys {
			if _, ok := blsKeys[key]; ok {
				found = true
			} else {
				blsKeys[key] = struct{}{}
				shard := new(big.Int).Mod(key.Big(), big.NewInt(int64(shardCount))).Int64()
				if slotsLimit == 0 || shardSlotsCount[shard] < slotsLimit {
					slotPubKeysLimited = append(slotPubKeysLimited, key)
				}
				shardSlotsCount[shard]++
			}
		}

		if found {
			continue
		}

		validatorStake := big.NewInt(0)
		for i := range validator.Delegations {
			validatorStake.Add(
				validatorStake, validator.Delegations[i].Amount,
			)
		}

		totalStaked.Add(totalStaked, validatorStake)

		essentials[validator.Address] = &effective.SlotOrder{
			Stake:       validatorStake,
			SpreadAmong: slotPubKeysLimited,
			Percentage:  tempZero,
		}
	}
	totalStakedDec := numeric.NewDecFromBigInt(totalStaked)

	for _, value := range essentials {
		value.Percentage = numeric.NewDecFromBigInt(value.Stake).Quo(totalStakedDec)
	}

	return essentials, nil
}

// IsEligibleForEPoSAuction ..
func IsEligibleForEPoSAuction(snapshot *staking.ValidatorSnapshot, validator *staking.ValidatorWrapper) bool {
	if validator.LastEpochInCommittee.Cmp(snapshot.Epoch) == 0 {
		computed := availability.ComputeCurrentSigning(snapshot.Validator, validator)
		if computed.IsBelowThreshold {
			return false
		}
	}
	switch validator.Status {
	case effective.Active:
		return true
	default:
		return false
	}
}

// ChainReader is a subset of Engine.Blockchain, just enough to do assignment
type ChainReader interface {
	ReadShardState(epoch *big.Int) (*shard.State, error)
	GetHeaderByHash(common.Hash) *block.Header
	Config() *params.ChainConfig
	CurrentHeader() *block.Header
}

// DataProvider ..
type DataProvider interface {
	StakingCandidatesReader
	ChainReader
}

type partialStakingEnabled struct{}

var (
	WithStakingEnabled           = partialStakingEnabled{}
	ErrComputeForEpochInPast     = errors.New("cannot compute for epoch in past")
)

// Pre-staking shard state computation logic
// Pre-staking shard state computation logic
// preStakingEnabledCommittee computes the pre-staking shard state
// preStakingEnabledCommittee computes the pre-staking shard state for IntelChain.
// preStakingEnabledCommittee computes the pre-staking shard state for IntelChain.
func preStakingEnabledCommittee(s shardingconfig.Instance) (*shard.State, error) {
    shardNum := int(s.NumShards())
    shardIntelchainNodes := s.NumIntelchainOperatedNodesPerShard()
    shardSize := s.NumNodesPerShard()
    itcAccounts := s.ItcAccounts()
    fnAccounts := s.FnAccounts()

    // Initialize the shard state
    shardState := &shard.State{}

    // Shard state needs to be sorted by shard ID
    for i := 0; i < shardNum; i++ {
        com := shard.Committee{ShardID: uint32(i)}

        // Add IntelChain operated nodes
        for j := 0; j < shardIntelchainNodes; j++ {
            index := i + j*shardNum // The initial account to use for genesis nodes
            pub := &bls_core.PublicKey{}

            // Deserialize BLS public key
            if err := pub.DeserializeHexStr(itcAccounts[index].BLSPublicKey); err != nil {
                fmt.Printf("Error deserializing BLS public key for IntelChain account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }

            pubKey := bls.SerializedPublicKey{}
            // Convert to serialized public key
            if err := pubKey.FromLibBLSPublicKey(pub); err != nil {
                fmt.Printf("Error converting BLS public key to serialized public key for IntelChain account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }

            // Update address parsing to handle new Bech32 HRP
            addr, err := common2.ParseAddr(itcAccounts[index].Address)
            if err != nil {
                fmt.Printf("Error parsing address for IntelChain account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }
            curNodeID := shard.Slot{
                EcdsaAddress: addr,
                BLSPublicKey: pubKey,
            }
            com.Slots = append(com.Slots, curNodeID)
        }

        // Add FN runner's key
        for j := shardIntelchainNodes; j < shardSize; j++ {
            index := i + (j-shardIntelchainNodes)*shardNum
            pub := &bls_core.PublicKey{}
            
            // Deserialize BLS public key
            if err := pub.DeserializeHexStr(fnAccounts[index].BLSPublicKey); err != nil {
                fmt.Printf("Error deserializing BLS public key for FN account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }

            pubKey := bls.SerializedPublicKey{}
            // Convert to serialized public key
            if err := pubKey.FromLibBLSPublicKey(pub); err != nil {
                fmt.Printf("Error converting BLS public key to serialized public key for FN account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }

            // Update address parsing for FN accounts
            addr, err := common2.ParseAddr(fnAccounts[index].Address)
            if err != nil {
                fmt.Printf("Error parsing address for FN account at index %d: %v\n", index, err)
                fmt.Println("Continuing to next account...")
                continue // Skip to the next account instead of returning
            }
            curNodeID := shard.Slot{
                EcdsaAddress: addr,
                BLSPublicKey: pubKey,
            }
            com.Slots = append(com.Slots, curNodeID)
        }

        // Append the committee for this shard to the state
        shardState.Shards = append(shardState.Shards, com)
    }

    fmt.Println("Completed processing all shards.")
    return shardState, nil
}




func eposStakedCommittee(
	epoch *big.Int, s shardingconfig.Instance, stakerReader DataProvider,
) (*shard.State, error) {
	shardCount := int(s.NumShards())
	shardState := &shard.State{}
	shardState.Shards = make([]shard.Committee, shardCount)
	hAccounts := s.ItcAccounts()
	shardIntelchainNodes := s.NumIntelchainOperatedNodesPerShard()

	for i := 0; i < shardCount; i++ {
		shardState.Shards[i] = shard.Committee{ShardID: uint32(i), Slots: shard.SlotList{}}
		for j := 0; j < shardIntelchainNodes; j++ {
			index := i + j*shardCount
			pub := &bls_core.PublicKey{}
			if err := pub.DeserializeHexStr(hAccounts[index].BLSPublicKey); err != nil {
				return nil, err
			}
			pubKey := bls.SerializedPublicKey{}
			if err := pubKey.FromLibBLSPublicKey(pub); err != nil {
				return nil, err
			}

			addr, err := common2.ParseAddr(hAccounts[index].Address)
			if err != nil {
				return nil, err
			}
			shardState.Shards[i].Slots = append(shardState.Shards[i].Slots, shard.Slot{
				EcdsaAddress: addr,
				BLSPublicKey: pubKey,
			})
		}
	}

	// TODO(audit): make sure external validator BLS key are also not duplicate to Intelchain's keys
	completedEPoSRound, err := NewEPoSRound(epoch, stakerReader, stakerReader.Config().IsEPoSBound35(epoch), s.SlotsLimit(), shardCount)

	if err != nil {
		return nil, err
	}

	shardBig := big.NewInt(int64(shardCount))
	for i := range completedEPoSRound.AuctionWinners {
		purchasedSlot := completedEPoSRound.AuctionWinners[i]
		shardID := int(new(big.Int).Mod(purchasedSlot.Key.Big(), shardBig).Int64())
		shardState.Shards[shardID].Slots = append(
			shardState.Shards[shardID].Slots, shard.Slot{
				EcdsaAddress:   purchasedSlot.Addr,
				BLSPublicKey:   purchasedSlot.Key,
				EffectiveStake: &purchasedSlot.EPoSStake,
			},
		)
	}

	if len(completedEPoSRound.AuctionWinners) == 0 {
		instance := shard.Schedule.InstanceForEpoch(epoch)
		preInstance := shard.Schedule.InstanceForEpoch(new(big.Int).Sub(epoch, big.NewInt(1)))
		isTestnet := nodeconfig.GetDefaultConfig().GetNetworkType() == nodeconfig.Testnet
		isShardReduction := preInstance.NumShards() != instance.NumShards()
		// If the shard-reduction happens, we cannot use the old committee.
		if isTestnet && isShardReduction {
			utils.Logger().Warn().Msg("No elected validators in the new epoch!!! But use the new committee due to Testnet Shard Reduction.")
			return shardState, nil
		}
		utils.Logger().Warn().Msg("No elected validators in the new epoch!!! Reuse old shard state.")
		return stakerReader.ReadShardState(big.NewInt(0).Sub(epoch, big.NewInt(1)))
	}
	return shardState, nil
}

// ReadFromDB is a wrapper on ReadShardState
func (def partialStakingEnabled) ReadFromDB(
	epoch *big.Int, reader DataProvider,
) (newSuperComm *shard.State, err error) {
	return reader.ReadShardState(epoch)
}

// Compute is single entry point for
// computing a new super committee, aka new shard state
func (def partialStakingEnabled) Compute(
	epoch *big.Int, stakerReader DataProvider,
) (newSuperComm *shard.State, err error) {
	preStaking := true
	if stakerReader != nil {
		config := stakerReader.Config()
		if config.IsStaking(epoch) {
			preStaking = false
		}
	}

	instance := shard.Schedule.InstanceForEpoch(epoch)
	if preStaking {
		// Pre-staking shard state doesn't need to set epoch (backward compatible)
		return preStakingEnabledCommittee(instance)
	}
	// Sanity check, can't compute against epochs in past
	if e := stakerReader.CurrentHeader().Epoch(); epoch.Cmp(e) == -1 {
		utils.Logger().Error().Uint64("header-epoch", e.Uint64()).
			Uint64("compute-epoch", epoch.Uint64()).
			Msg("Tried to compute committee for epoch in past")
		return nil, ErrComputeForEpochInPast
	}
	utils.AnalysisStart("computeEPoSStakedCommittee")
	shardState, err := eposStakedCommittee(epoch, instance, stakerReader)
	utils.AnalysisEnd("computeEPoSStakedCommittee")

	if err != nil {
		return nil, err
	}

	// Set the epoch of shard state
	shardState.Epoch = big.NewInt(0).Set(epoch)
	utils.Logger().Info().
		Uint64("computed-for-epoch", epoch.Uint64()).
		Msg("computed new super committee")
	return shardState, nil
}
