package itc

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/zennittians/intelchain/block"
	"github.com/zennittians/intelchain/consensus/quorum"
	"github.com/zennittians/intelchain/core/rawdb"
	"github.com/zennittians/intelchain/core/state"
	"github.com/zennittians/intelchain/core/types"
	"github.com/zennittians/intelchain/eth/rpc"
	"github.com/zennittians/intelchain/internal/chain"
	internalCommon "github.com/zennittians/intelchain/internal/common"
	"github.com/zennittians/intelchain/numeric"
	commonRPC "github.com/zennittians/intelchain/rpc/intelchain/common"
	"github.com/zennittians/intelchain/shard"
	"github.com/zennittians/intelchain/shard/committee"
	"github.com/zennittians/intelchain/staking/availability"
	"github.com/zennittians/intelchain/staking/effective"
	staking "github.com/zennittians/intelchain/staking/types"
)

var (
	zero    = numeric.ZeroDec()
	bigZero = big.NewInt(0)
)

func (itc *Intelchain) readAndUpdateRawStakes(
	epoch *big.Int,
	decider quorum.Decider,
	comm shard.Committee,
	rawStakes []effective.SlotPurchase,
	validatorSpreads map[common.Address]numeric.Dec,
) []effective.SlotPurchase {
	for i := range comm.Slots {
		slot := comm.Slots[i]
		slotAddr := slot.EcdsaAddress
		slotKey := slot.BLSPublicKey
		spread, ok := validatorSpreads[slotAddr]
		if !ok {
			snapshot, err := itc.BlockChain.ReadValidatorSnapshotAtEpoch(epoch, slotAddr)
			if err != nil {
				continue
			}
			spread = snapshot.RawStakePerSlot()
			validatorSpreads[slotAddr] = spread
		}

		commonRPC.SetRawStake(decider, slotKey, spread)
		// add entry to array for median calculation
		rawStakes = append(rawStakes, effective.SlotPurchase{
			Addr:      slotAddr,
			Key:       slotKey,
			RawStake:  spread,
			EPoSStake: spread,
		})
	}
	return rawStakes
}

func (itc *Intelchain) getSuperCommittees() (*quorum.Transition, error) {
	nowE := itc.BlockChain.CurrentHeader().Epoch()

	if itc.BlockChain.CurrentHeader().IsLastBlockInEpoch() {
		// current epoch is current header epoch + 1 if the header was last block of prev epoch
		nowE = new(big.Int).Add(nowE, common.Big1)
	}
	thenE := new(big.Int).Sub(nowE, common.Big1)

	var (
		nowCommittee, prevCommittee *shard.State
		err                         error
	)
	nowCommittee, err = itc.BlockChain.ReadShardState(nowE)
	if err != nil {
		return nil, err
	}
	prevCommittee, err = itc.BlockChain.ReadShardState(thenE)
	if err != nil {
		return nil, err
	}

	stakedSlotsNow, stakedSlotsThen :=
		shard.ExternalSlotsAvailableForEpoch(nowE),
		shard.ExternalSlotsAvailableForEpoch(thenE)

	then, now :=
		quorum.NewRegistry(stakedSlotsThen, int(thenE.Int64())),
		quorum.NewRegistry(stakedSlotsNow, int(nowE.Int64()))

	rawStakes := []effective.SlotPurchase{}
	validatorSpreads := map[common.Address]numeric.Dec{}
	for _, comm := range prevCommittee.Shards {
		decider := quorum.NewDecider(quorum.SuperMajorityStake, comm.ShardID)
		// before staking skip computing
		if itc.BlockChain.Config().IsStaking(prevCommittee.Epoch) {
			if _, err := decider.SetVoters(&comm, prevCommittee.Epoch); err != nil {
				return nil, err
			}
		}
		rawStakes = itc.readAndUpdateRawStakes(thenE, decider, comm, rawStakes, validatorSpreads)
		then.Deciders[fmt.Sprintf("shard-%d", comm.ShardID)] = decider
	}
	then.MedianStake = effective.Median(rawStakes)

	rawStakes = []effective.SlotPurchase{}
	validatorSpreads = map[common.Address]numeric.Dec{}
	for _, comm := range nowCommittee.Shards {
		decider := quorum.NewDecider(quorum.SuperMajorityStake, comm.ShardID)
		if _, err := decider.SetVoters(&comm, nowCommittee.Epoch); err != nil {
			return nil, errors.Wrapf(
				err,
				"committee is only available from staking epoch: %v, current epoch: %v",
				itc.BlockChain.Config().StakingEpoch,
				itc.BlockChain.CurrentHeader().Epoch(),
			)
		}
		rawStakes = itc.readAndUpdateRawStakes(nowE, decider, comm, rawStakes, validatorSpreads)
		now.Deciders[fmt.Sprintf("shard-%d", comm.ShardID)] = decider
	}
	now.MedianStake = effective.Median(rawStakes)

	return &quorum.Transition{Previous: then, Current: now}, nil
}

// IsStakingEpoch ...
func (itc *Intelchain) IsStakingEpoch(epoch *big.Int) bool {
	return itc.BlockChain.Config().IsStaking(epoch)
}

// IsPreStakingEpoch ...
func (itc *Intelchain) IsPreStakingEpoch(epoch *big.Int) bool {
	return itc.BlockChain.Config().IsPreStaking(epoch)
}

// IsNoEarlyUnlockEpoch ...
func (itc *Intelchain) IsNoEarlyUnlockEpoch(epoch *big.Int) bool {
	return itc.BlockChain.Config().IsNoEarlyUnlock(epoch)
}

// IsMaxRate ...
func (itc *Intelchain) IsMaxRate(epoch *big.Int) bool {
	return itc.BlockChain.Config().IsMaxRate(epoch)
}

// IsCommitteeSelectionBlock checks if the given block is the committee selection block
func (itc *Intelchain) IsCommitteeSelectionBlock(header *block.Header) bool {
	return chain.IsCommitteeSelectionBlock(itc.BlockChain, header)
}

// GetDelegationLockingPeriodInEpoch ...
func (itc *Intelchain) GetDelegationLockingPeriodInEpoch(epoch *big.Int) int {
	return chain.GetLockPeriodInEpoch(itc.BlockChain, epoch)
}

// SendStakingTx adds a staking transaction
func (itc *Intelchain) SendStakingTx(ctx context.Context, signedStakingTx *staking.StakingTransaction) error {
	stx, _, _, _ := rawdb.ReadStakingTransaction(itc.chainDb, signedStakingTx.Hash())
	if stx == nil {
		return itc.NodeAPI.AddPendingStakingTransaction(signedStakingTx)
	}
	return ErrFinalizedTransaction
}

// GetStakingTransactionsHistory returns list of staking transactions hashes of address.
func (itc *Intelchain) GetStakingTransactionsHistory(address, txType, order string) ([]common.Hash, error) {
	return itc.NodeAPI.GetStakingTransactionsHistory(address, txType, order)
}

// GetStakingTransactionsCount returns the number of staking transactions of address.
func (itc *Intelchain) GetStakingTransactionsCount(address, txType string) (uint64, error) {
	return itc.NodeAPI.GetStakingTransactionsCount(address, txType)
}

// GetSuperCommittees ..
func (itc *Intelchain) GetSuperCommittees() (*quorum.Transition, error) {
	nowE := itc.BlockChain.CurrentHeader().Epoch()
	key := fmt.Sprintf("sc-%s", nowE.String())

	res, err := itc.SingleFlightRequest(
		key, func() (interface{}, error) {
			thenE := new(big.Int).Sub(nowE, common.Big1)
			thenKey := fmt.Sprintf("sc-%s", thenE.String())
			itc.group.Forget(thenKey)
			return itc.getSuperCommittees()
		})
	if err != nil {
		return nil, err
	}
	return res.(*quorum.Transition), err
}

// GetValidators returns validators for a particular epoch.
func (itc *Intelchain) GetValidators(epoch *big.Int) (*shard.Committee, error) {
	state, err := itc.BlockChain.ReadShardState(epoch)
	if err != nil {
		return nil, err
	}
	for _, cmt := range state.Shards {
		if cmt.ShardID == itc.ShardID {
			return &cmt, nil
		}
	}
	return nil, nil
}

// GetValidatorSelfDelegation returns the amount of staking after applying all delegated stakes
func (itc *Intelchain) GetValidatorSelfDelegation(addr common.Address) *big.Int {
	wrapper, err := itc.BlockChain.ReadValidatorInformation(addr)
	if err != nil || wrapper == nil {
		return nil
	}
	if len(wrapper.Delegations) == 0 {
		return nil
	}
	return wrapper.Delegations[0].Amount
}

// GetElectedValidatorAddresses returns the address of elected validators for current epoch
func (itc *Intelchain) GetElectedValidatorAddresses() []common.Address {
	list, _ := itc.BlockChain.ReadShardState(itc.BlockChain.CurrentBlock().Epoch())
	return list.StakedValidators().Addrs
}

// GetAllValidatorAddresses returns the up to date validator candidates for next epoch
func (itc *Intelchain) GetAllValidatorAddresses() []common.Address {
	return itc.BlockChain.ValidatorCandidates()
}

func (itc *Intelchain) GetValidatorsStakeByBlockNumber(
	block *types.Block,
) (map[string]*big.Int, error) {
	if cachedReward, ok := itc.stakeByBlockNumberCache.Get(block.Hash()); ok {
		return cachedReward.(map[string]*big.Int), nil
	}
	validatorAddresses := itc.GetAllValidatorAddresses()
	stakes := make(map[string]*big.Int, len(validatorAddresses))
	for _, validatorAddress := range validatorAddresses {
		wrapper, err := itc.BlockChain.ReadValidatorInformationAtRoot(validatorAddress, block.Root())
		if err != nil {
			if errors.Cause(err) != state.ErrAddressNotPresent {
				return nil, errors.Errorf(
					"cannot fetch information for validator %s at block %d due to %s",
					validatorAddress.Hex(),
					block.Number(),
					err,
				)
			} else {
				// `validatorAddress` was not a validator back then
				continue
			}
		}
		stakes[validatorAddress.Hex()] = wrapper.TotalDelegation()
	}
	itc.stakeByBlockNumberCache.Add(block.Hash(), stakes)
	return stakes, nil
}

var (
	epochBlocksMap = map[common.Address]map[uint64]staking.EpochSigningEntry{}
	mapLock        = sync.Mutex{}
)

func (itc *Intelchain) getEpochSigning(epoch *big.Int, addr common.Address) (staking.EpochSigningEntry, error) {
	entry := staking.EpochSigningEntry{}
	mapLock.Lock()
	defer mapLock.Unlock()
	if validatorMap, ok := epochBlocksMap[addr]; ok {
		if val, ok := validatorMap[epoch.Uint64()]; ok {
			return val, nil
		}
	}

	snapshot, err := itc.BlockChain.ReadValidatorSnapshotAtEpoch(epoch, addr)
	if err != nil {
		return entry, err
	}

	// the signing information is for the previous epoch
	prevEpoch := big.NewInt(0).Sub(epoch, common.Big1)
	entry.Epoch = prevEpoch
	entry.Blocks = snapshot.Validator.Counters

	// subtract previous epoch counters if exists
	prevEpochSnap, err := itc.BlockChain.ReadValidatorSnapshotAtEpoch(prevEpoch, addr)
	if err == nil {
		entry.Blocks.NumBlocksSigned.Sub(
			entry.Blocks.NumBlocksSigned,
			prevEpochSnap.Validator.Counters.NumBlocksSigned,
		)
		entry.Blocks.NumBlocksToSign.Sub(
			entry.Blocks.NumBlocksToSign,
			prevEpochSnap.Validator.Counters.NumBlocksToSign,
		)
	}

	// update map when adding new entry, also remove an entry beyond last 30
	if _, ok := epochBlocksMap[addr]; !ok {
		epochBlocksMap[addr] = map[uint64]staking.EpochSigningEntry{}
	}
	epochBlocksMap[addr][epoch.Uint64()] = entry
	epochMinus30 := big.NewInt(0).Sub(epoch, big.NewInt(staking.SigningHistoryLength))
	delete(epochBlocksMap[addr], epochMinus30.Uint64())

	return entry, nil
}

// GetValidatorInformation returns the information of validator
func (itc *Intelchain) GetValidatorInformation(
	addr common.Address, block *types.Block,
) (*staking.ValidatorRPCEnhanced, error) {
	bc := itc.BlockChain
	wrapper, err := bc.ReadValidatorInformationAtRoot(addr, block.Root())
	if err != nil {
		s, _ := internalCommon.AddressToBech32(addr)
		return nil, errors.Wrapf(err, "not found address in current state %s", s)
	}

	now := block.Epoch()
	// At the last block of epoch, block epoch is e while val.LastEpochInCommittee
	// is already updated to e+1. So need the >= check rather than ==
	inCommittee := wrapper.LastEpochInCommittee.Cmp(now) >= 0
	defaultReply := &staking.ValidatorRPCEnhanced{
		CurrentlyInCommittee: inCommittee,
		Wrapper:              *wrapper,
		Performance:          nil,
		ComputedMetrics:      nil,
		TotalDelegated:       wrapper.TotalDelegation(),
		EPoSStatus: effective.ValidatorStatus(
			inCommittee, wrapper.Status,
		).String(),
		EPoSWinningStake: nil,
		BootedStatus:     nil,
		ActiveStatus:     wrapper.Validator.Status.String(),
		Lifetime: &staking.AccumulatedOverLifetime{
			BlockReward: wrapper.BlockReward,
			Signing:     wrapper.Counters,
			APR:         zero,
		},
	}

	snapshot, err := bc.ReadValidatorSnapshotAtEpoch(
		now, addr,
	)

	if err != nil {
		return defaultReply, nil
	}

	computed := availability.ComputeCurrentSigning(
		snapshot.Validator, wrapper,
	)

	lastBlockOfEpoch := shard.Schedule.EpochLastBlock(itc.BeaconChain.CurrentBlock().Header().Epoch().Uint64())

	computed.BlocksLeftInEpoch = lastBlockOfEpoch - itc.BeaconChain.CurrentBlock().Header().Number().Uint64()

	if defaultReply.CurrentlyInCommittee {
		defaultReply.Performance = &staking.CurrentEpochPerformance{
			CurrentSigningPercentage: *computed,
			Epoch:                    itc.BeaconChain.CurrentBlock().Header().Epoch().Uint64(),
			Block:                    itc.BeaconChain.CurrentBlock().Header().Number().Uint64(),
		}
	}

	stats, err := bc.ReadValidatorStats(addr)
	if err != nil {
		// when validator has no stats, default boot-status to not booted
		notBooted := effective.NotBooted.String()
		defaultReply.BootedStatus = &notBooted
		return defaultReply, nil
	}

	latestAPR := numeric.ZeroDec()
	l := len(stats.APRs)
	if l > 0 {
		latestAPR = stats.APRs[l-1].Value
	}
	defaultReply.Lifetime.APR = latestAPR
	defaultReply.Lifetime.EpochAPRs = stats.APRs

	// average apr cache keys
	// key := fmt.Sprintf("apr-%s-%d", addr.Hex(), now.Uint64())
	// prevKey := fmt.Sprintf("apr-%s-%d", addr.Hex(), now.Uint64()-1)

	// delete entry for previous epoch
	// b.apiCache.Forget(prevKey)

	// calculate last APRHistoryLength epochs for averaging APR
	// epochFrom := bc.GasPriceConfig().StakingEpoch
	// nowMinus := big.NewInt(0).Sub(now, big.NewInt(staking.APRHistoryLength))
	// if nowMinus.Cmp(epochFrom) > 0 {
	// 	epochFrom = nowMinus
	// }

	// if len(stats.APRs) > 0 && stats.APRs[0].Epoch.Cmp(epochFrom) > 0 {
	// 	epochFrom = stats.APRs[0].Epoch
	// }

	// epochToAPRs := map[int64]numeric.Dec{}
	// for i := 0; i < len(stats.APRs); i++ {
	// 	entry := stats.APRs[i]
	// 	epochToAPRs[entry.Epoch.Int64()] = entry.Value
	// }

	// at this point, validator is active and has apr's for the recent 100 epochs
	// compute average apr over history
	// if avgAPR, err := b.SingleFlightRequest(
	// 	key, func() (interface{}, error) {
	// 		total := numeric.ZeroDec()
	// 		count := 0
	// 		for i := epochFrom.Int64(); i < now.Int64(); i++ {
	// 			if apr, ok := epochToAPRs[i]; ok {
	// 				total = total.Add(apr)
	// 			}
	// 			count++
	// 		}
	// 		if count == 0 {
	// 			return nil, errors.New("no apr snapshots available")
	// 		}
	// 		return total.QuoInt64(int64(count)), nil
	// 	},
	// ); err != nil {
	// 	// could not compute average apr from snapshot
	// 	// assign the latest apr available from stats
	// 	defaultReply.Lifetime.APR = numeric.ZeroDec()
	// } else {
	// 	defaultReply.Lifetime.APR = avgAPR.(numeric.Dec)
	// }

	epochBlocks := []staking.EpochSigningEntry{}
	epochFrom := bc.Config().StakingEpoch
	nowMinus := big.NewInt(0).Sub(now, big.NewInt(staking.SigningHistoryLength))
	if nowMinus.Cmp(epochFrom) > 0 {
		epochFrom = nowMinus
	}
	for i := now.Int64(); i > epochFrom.Int64(); i-- {
		epoch := big.NewInt(i)
		entry, err := itc.getEpochSigning(epoch, addr)
		if err != nil {
			break
		}
		epochBlocks = append(epochBlocks, entry)
	}
	defaultReply.Lifetime.EpochBlocks = epochBlocks

	if defaultReply.CurrentlyInCommittee {
		defaultReply.ComputedMetrics = stats
		defaultReply.EPoSWinningStake = &stats.TotalEffectiveStake
	}

	if !defaultReply.CurrentlyInCommittee {
		reason := stats.BootedStatus.String()
		defaultReply.BootedStatus = &reason
	}

	return defaultReply, nil
}

// GetMedianRawStakeSnapshot ..
func (itc *Intelchain) GetMedianRawStakeSnapshot() (
	*committee.CompletedEPoSRound, error,
) {
	blockNum := itc.CurrentBlock().NumberU64()
	key := fmt.Sprintf("median-%d", blockNum)

	// delete cache for previous block
	prevKey := fmt.Sprintf("median-%d", blockNum-1)
	itc.group.Forget(prevKey)

	res, err := itc.SingleFlightRequest(
		key,
		func() (interface{}, error) {
			// Compute for next epoch
			epoch := big.NewInt(0).Add(itc.CurrentBlock().Epoch(), big.NewInt(1))
			instance := shard.Schedule.InstanceForEpoch(epoch)
			return committee.NewEPoSRound(epoch, itc.BlockChain, itc.BlockChain.Config().IsEPoSBound35(epoch), instance.SlotsLimit(), int(instance.NumShards()))
		},
	)
	if err != nil {
		return nil, err
	}
	return res.(*committee.CompletedEPoSRound), nil
}

// GetDelegationsByValidator returns all delegation information of a validator
func (itc *Intelchain) GetDelegationsByValidator(validator common.Address) []staking.Delegation {
	wrapper, err := itc.BlockChain.ReadValidatorInformation(validator)
	if err != nil || wrapper == nil {
		return nil
	}
	return wrapper.Delegations
}

// GetDelegationsByValidatorAtBlock returns all delegation information of a validator at the given block
func (itc *Intelchain) GetDelegationsByValidatorAtBlock(
	validator common.Address, block *types.Block,
) []staking.Delegation {
	wrapper, err := itc.BlockChain.ReadValidatorInformationAtRoot(validator, block.Root())
	if err != nil || wrapper == nil {
		return nil
	}
	return wrapper.Delegations
}

// GetDelegationsByDelegator returns all delegation information of a delegator
func (itc *Intelchain) GetDelegationsByDelegator(
	delegator common.Address,
) ([]common.Address, []*staking.Delegation) {
	block := itc.BlockChain.CurrentBlock()
	return itc.GetDelegationsByDelegatorByBlock(delegator, block)
}

// GetDelegationsByDelegatorByBlock returns all delegation information of a delegator
func (itc *Intelchain) GetDelegationsByDelegatorByBlock(
	delegator common.Address, block *types.Block,
) ([]common.Address, []*staking.Delegation) {
	delegationIndexes, err := itc.BlockChain.
		ReadDelegationsByDelegatorAt(delegator, block.Number())
	if err != nil {
		return nil, nil
	}

	addresses := make([]common.Address, 0, len(delegationIndexes))
	delegations := make([]*staking.Delegation, 0, len(delegationIndexes))

	for i := range delegationIndexes {
		wrapper, err := itc.BlockChain.ReadValidatorInformationAtRoot(
			delegationIndexes[i].ValidatorAddress, block.Root(),
		)
		if err != nil || wrapper == nil {
			return nil, nil
		}

		if uint64(len(wrapper.Delegations)) > delegationIndexes[i].Index {
			delegations = append(delegations, &wrapper.Delegations[delegationIndexes[i].Index])
		} else {
			delegations = append(delegations, nil)
		}
		addresses = append(addresses, delegationIndexes[i].ValidatorAddress)
	}
	return addresses, delegations
}

// UndelegationPayouts ..
type UndelegationPayouts struct {
	Data map[common.Address]map[common.Address]*big.Int
}

func NewUndelegationPayouts() *UndelegationPayouts {
	return &UndelegationPayouts{
		Data: make(map[common.Address]map[common.Address]*big.Int),
	}
}

func (u *UndelegationPayouts) SetPayoutByDelegatorAddrAndValidatorAddr(
	delegator, validator common.Address, amount *big.Int,
) {
	if u.Data[delegator] == nil {
		u.Data[delegator] = make(map[common.Address]*big.Int)
	}

	if totalPayout, ok := u.Data[delegator][validator]; ok {
		u.Data[delegator][validator] = new(big.Int).Add(totalPayout, amount)
	} else {
		u.Data[delegator][validator] = amount
	}
}

// GetUndelegationPayouts returns the undelegation payouts for each delegator
//
// Due to in-memory caching, it is possible to get undelegation payouts for a state / epoch
// that has been pruned but have it be lost (and unable to recompute) after the node restarts.
// This not a problem if a full (archival) DB is used.
func (itc *Intelchain) GetUndelegationPayouts(
	ctx context.Context, epoch *big.Int,
) (*UndelegationPayouts, error) {
	if !itc.IsPreStakingEpoch(epoch) {
		return nil, fmt.Errorf("not pre-staking epoch or later")
	}

	payouts, ok := itc.undelegationPayoutsCache.Get(epoch.Uint64())
	if ok {
		return payouts.(*UndelegationPayouts), nil
	}
	undelegationPayouts := NewUndelegationPayouts()
	// require second to last block as saved undelegations are AFTER undelegations are payed out
	blockNumber := shard.Schedule.EpochLastBlock(epoch.Uint64()) - 1
	undelegationPayoutBlock, err := itc.BlockByNumber(ctx, rpc.BlockNumber(blockNumber))
	if err != nil || undelegationPayoutBlock == nil {
		// Block not found, so no undelegationPayouts (not an error)
		return undelegationPayouts, nil
	}

	isMaxRate := itc.IsMaxRate(epoch)
	lockingPeriod := itc.GetDelegationLockingPeriodInEpoch(undelegationPayoutBlock.Epoch())
	for _, validator := range itc.GetAllValidatorAddresses() {
		wrapper, err := itc.BlockChain.ReadValidatorInformationAtRoot(validator, undelegationPayoutBlock.Root())
		if err != nil || wrapper == nil {
			continue // Not a validator at this epoch or unable to fetch validator info because of pruned state.
		}
		noEarlyUnlock := itc.IsNoEarlyUnlockEpoch(epoch)
		for _, delegation := range wrapper.Delegations {
			withdraw := delegation.RemoveUnlockedUndelegations(epoch, wrapper.LastEpochInCommittee, lockingPeriod, noEarlyUnlock, isMaxRate)
			if withdraw.Cmp(bigZero) == 1 {
				undelegationPayouts.SetPayoutByDelegatorAddrAndValidatorAddr(validator, delegation.DelegatorAddress, withdraw)
			}
		}
	}

	itc.undelegationPayoutsCache.Add(epoch.Uint64(), undelegationPayouts)
	return undelegationPayouts, nil
}

// GetTotalStakingSnapshot ..
func (itc *Intelchain) GetTotalStakingSnapshot() *big.Int {
	if stake := itc.totalStakeCache.pop(itc.CurrentBlock().NumberU64()); stake != nil {
		return stake
	}
	currHeight := itc.CurrentBlock().NumberU64()
	candidates := itc.BlockChain.ValidatorCandidates()
	if len(candidates) == 0 {
		stake := big.NewInt(0)
		itc.totalStakeCache.push(currHeight, stake)
		return stake
	}
	stakes := big.NewInt(0)
	for i := range candidates {
		snapshot, _ := itc.BlockChain.ReadValidatorSnapshot(candidates[i])
		validator, _ := itc.BlockChain.ReadValidatorInformation(candidates[i])
		if !committee.IsEligibleForEPoSAuction(
			snapshot, validator,
		) {
			continue
		}
		for i := range validator.Delegations {
			stakes.Add(stakes, validator.Delegations[i].Amount)
		}
	}
	itc.totalStakeCache.push(currHeight, stakes)
	return stakes
}

// GetCurrentStakingErrorSink ..
func (itc *Intelchain) GetCurrentStakingErrorSink() types.TransactionErrorReports {
	return itc.NodeAPI.ReportStakingErrorSink()
}

// totalStakeCache ..
type totalStakeCache struct {
	sync.Mutex
	cachedBlockHeight uint64
	stake             *big.Int
	// duration is in blocks
	duration uint64
}

// newTotalStakeCache ..
func newTotalStakeCache(duration uint64) *totalStakeCache {
	return &totalStakeCache{
		cachedBlockHeight: 0,
		stake:             nil,
		duration:          duration,
	}
}

func (c *totalStakeCache) push(currBlockHeight uint64, stake *big.Int) {
	c.Lock()
	defer c.Unlock()
	c.cachedBlockHeight = currBlockHeight
	c.stake = stake
}

func (c *totalStakeCache) pop(currBlockHeight uint64) *big.Int {
	if currBlockHeight > c.cachedBlockHeight+c.duration {
		return nil
	}
	return c.stake
}
