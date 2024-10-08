package votepower

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sort"

	"github.com/zennittians/intelchain/shard"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	bls_core "github.com/zennittians/bls/ffi/go/bls"
	"github.com/zennittians/intelchain/crypto/bls"
	common2 "github.com/zennittians/intelchain/internal/common"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/numeric"
)

var (
	// ErrVotingPowerNotEqualOne ..
	ErrVotingPowerNotEqualOne = errors.New("voting power not equal to one")
)

// Ballot is a vote cast by a validator
type Ballot struct {
	SignerPubKeys   []bls.SerializedPublicKey `json:"bls-public-keys"`
	BlockHeaderHash common.Hash               `json:"block-header-hash"`
	Signature       []byte                    `json:"bls-signature"`
	Height          uint64                    `json:"block-height"`
	ViewID          uint64                    `json:"view-id"`
}

// MarshalJSON ..
func (b Ballot) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		A string `json:"bls-public-keys"`
		B string `json:"block-header-hash"`
		C string `json:"bls-signature"`
		E uint64 `json:"block-height"`
		F uint64 `json:"view-id"`
	}{
		fmt.Sprint(b.SignerPubKeys),
		b.BlockHeaderHash.Hex(),
		hex.EncodeToString(b.Signature),
		b.Height,
		b.ViewID,
	})
}

// Round is a round of voting in any FBFT phase
type Round struct {
	AggregatedVote *bls_core.Sign
	BallotBox      map[bls.SerializedPublicKey]*Ballot
}

func (b Ballot) String() string {
	data, _ := json.Marshal(b)
	return string(data)
}

// NewRound ..
func NewRound() *Round {
	return &Round{
		AggregatedVote: &bls_core.Sign{},
		BallotBox:      map[bls.SerializedPublicKey]*Ballot{},
	}
}

// PureStakedVote ..
type PureStakedVote struct {
	EarningAccount common.Address          `json:"earning-account"`
	Identity       bls.SerializedPublicKey `json:"bls-public-key"`
	GroupPercent   numeric.Dec             `json:"group-percent"`
	EffectiveStake numeric.Dec             `json:"effective-stake"`
	RawStake       numeric.Dec             `json:"raw-stake"`
}

// AccommodateIntelchainVote ..
type AccommodateIntelchainVote struct {
	PureStakedVote
	IsIntelchainNode bool        `json:"-"`
	OverallPercent   numeric.Dec `json:"overall-percent"`
}

// String ..
func (v AccommodateIntelchainVote) String() string {
	s, _ := json.Marshal(v)
	return string(s)
}

// Roster ..
type Roster struct {
	Voters       map[bls.SerializedPublicKey]*AccommodateIntelchainVote
	ShardID      uint32
	OrderedSlots []bls.SerializedPublicKey

	OurVotingPowerTotalPercentage   numeric.Dec
	TheirVotingPowerTotalPercentage numeric.Dec
	TotalEffectiveStake             numeric.Dec
	ITCSlotCount                    int64
}

func (r Roster) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}

// VoteOnSubcomittee ..
type VoteOnSubcomittee struct {
	AccommodateIntelchainVote
	ShardID uint32
}

// MarshalJSON ..
func (v VoteOnSubcomittee) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PureStakedVote
		EarningAccount string      `json:"earning-account"`
		OverallPercent numeric.Dec `json:"overall-percent"`
		ShardID        uint32      `json:"shard-id"`
	}{
		v.PureStakedVote,
		common2.MustAddressToBech32(v.EarningAccount),
		v.OverallPercent,
		v.ShardID,
	})
}

// AggregateRosters ..
func AggregateRosters(
	rosters []*Roster,
) map[common.Address][]VoteOnSubcomittee {
	result := map[common.Address][]VoteOnSubcomittee{}
	sort.SliceStable(rosters, func(i, j int) bool {
		return rosters[i].ShardID < rosters[j].ShardID
	})

	for _, roster := range rosters {
		for _, voteCard := range roster.Voters {
			if !voteCard.IsIntelchainNode {
				voterID := VoteOnSubcomittee{
					AccommodateIntelchainVote: *voteCard,
					ShardID:                   roster.ShardID,
				}
				result[voteCard.EarningAccount] = append(
					result[voteCard.EarningAccount], voterID,
				)
			}
		}
	}

	return result
}

// Compute creates a new roster based off the shard.SlotList
func Compute(subComm *shard.Committee, epoch *big.Int) (*Roster, error) {
	if epoch == nil {
		return nil, errors.New("nil epoch for roster compute")
	}
	roster, staked := NewRoster(subComm.ShardID), subComm.Slots

	for i := range staked {
		if e := staked[i].EffectiveStake; e != nil {
			roster.TotalEffectiveStake = roster.TotalEffectiveStake.Add(*e)
		} else {
			roster.ITCSlotCount++
		}
	}

	asDecITCSlotCount := numeric.NewDec(roster.ITCSlotCount)
	// TODO Check for duplicate BLS Keys
	ourPercentage := numeric.ZeroDec()
	theirPercentage := numeric.ZeroDec()
	var lastStakedVoter *AccommodateIntelchainVote

	intelchainPercent := shard.Schedule.InstanceForEpoch(epoch).IntelchainVotePercent()
	externalPercent := shard.Schedule.InstanceForEpoch(epoch).ExternalVotePercent()

	for i := range staked {
		member := AccommodateIntelchainVote{
			PureStakedVote: PureStakedVote{
				EarningAccount: staked[i].EcdsaAddress,
				Identity:       staked[i].BLSPublicKey,
				GroupPercent:   numeric.ZeroDec(),
				EffectiveStake: numeric.ZeroDec(),
				RawStake:       numeric.ZeroDec(),
			},
			OverallPercent:   numeric.ZeroDec(),
			IsIntelchainNode: false,
		}

		// Real Staker
		if e := staked[i].EffectiveStake; e != nil {
			member.EffectiveStake = member.EffectiveStake.Add(*e)
			member.GroupPercent = e.Quo(roster.TotalEffectiveStake)
			member.OverallPercent = member.GroupPercent.Mul(externalPercent)
			theirPercentage = theirPercentage.Add(member.OverallPercent)
			lastStakedVoter = &member
		} else { // Our node
			member.IsIntelchainNode = true
			member.OverallPercent = intelchainPercent.Quo(asDecITCSlotCount)
			member.GroupPercent = member.OverallPercent.Quo(intelchainPercent)
			ourPercentage = ourPercentage.Add(member.OverallPercent)
		}

		if _, ok := roster.Voters[staked[i].BLSPublicKey]; !ok {
			roster.Voters[staked[i].BLSPublicKey] = &member
		} else {
			utils.Logger().Debug().Str("blsKey", staked[i].BLSPublicKey.Hex()).Msg("Duplicate BLS key found")
		}
	}

	{
		// NOTE Enforce voting power sums to one,
		// give diff (expect tiny amt) to last staked voter
		if diff := numeric.OneDec().Sub(
			ourPercentage.Add(theirPercentage),
		); !diff.IsZero() && lastStakedVoter != nil {
			lastStakedVoter.OverallPercent =
				lastStakedVoter.OverallPercent.Add(diff)
			theirPercentage = theirPercentage.Add(diff)
		}

		if lastStakedVoter != nil &&
			!ourPercentage.Add(theirPercentage).Equal(numeric.OneDec()) {
			return nil, ErrVotingPowerNotEqualOne
		}
	}

	roster.OurVotingPowerTotalPercentage = ourPercentage
	roster.TheirVotingPowerTotalPercentage = theirPercentage
	for _, slot := range subComm.Slots {
		roster.OrderedSlots = append(roster.OrderedSlots, slot.BLSPublicKey)
	}
	return roster, nil
}

// NewRoster ..
func NewRoster(shardID uint32) *Roster {
	m := map[bls.SerializedPublicKey]*AccommodateIntelchainVote{}
	return &Roster{
		Voters:  m,
		ShardID: shardID,

		OurVotingPowerTotalPercentage:   numeric.ZeroDec(),
		TheirVotingPowerTotalPercentage: numeric.ZeroDec(),
		TotalEffectiveStake:             numeric.ZeroDec(),
	}
}

// VotePowerByMask return the vote power with the given BLS mask. The result is a number between 0 and 1.
func (r *Roster) VotePowerByMask(mask *bls.Mask) numeric.Dec {
	res := numeric.ZeroDec()

	for key, index := range mask.PublicsIndex {
		if enabled, err := mask.IndexEnabled(index); err != nil || !enabled {
			continue
		}
		voter, ok := r.Voters[key]
		if !ok {
			continue
		}
		res = res.Add(voter.OverallPercent)
	}
	return res
}
