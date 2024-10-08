package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/zennittians/intelchain/core"
	itctypes "github.com/zennittians/intelchain/core/types"
	"github.com/zennittians/intelchain/rosetta/common"
)

// containsSideEffectTransaction checks if the block contains any side effect operations to report.
func (s *BlockAPI) containsSideEffectTransaction(
	ctx context.Context, blk *itctypes.Block,
) bool {
	if blk == nil {
		return false
	}
	return s.itc.IsCommitteeSelectionBlock(blk.Header()) || !s.itc.IsStakingEpoch(blk.Epoch()) || blk.NumberU64() == 0
}

const (
	// SideEffectTransactionSuffix is use in the transaction identifier for each block that contains
	// side-effect operations.
	SideEffectTransactionSuffix = "side_effect"
	blockHashStrLen             = 64
)

// getSideEffectTransactionIdentifier fetches 'transaction identifier' for side effect operations
// for a  given block.
// Side effects are genesis funds, pre-staking era block rewards, and undelegation payouts.
// Must include block hash to guarantee uniqueness of tx identifiers.
func getSideEffectTransactionIdentifier(
	blockHash ethcommon.Hash,
) *types.TransactionIdentifier {
	return &types.TransactionIdentifier{
		Hash: fmt.Sprintf("%v_%v",
			blockHash.String(), SideEffectTransactionSuffix,
		),
	}
}

// unpackSideEffectTransactionIdentifier returns the blockHash if the txID is formatted correctly.
func unpackSideEffectTransactionIdentifier(
	txID *types.TransactionIdentifier,
) (ethcommon.Hash, *types.Error) {
	hash := txID.Hash
	hash = strings.TrimPrefix(hash, "0x")
	hash = strings.TrimPrefix(hash, "0X")
	if len(hash) <= blockHashStrLen || string(hash[blockHashStrLen]) != "_" ||
		hash[blockHashStrLen+1:] != SideEffectTransactionSuffix {
		return ethcommon.Hash{}, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "unknown side effect transaction ID format",
		})
	}
	blkHash := ethcommon.HexToHash(hash[:blockHashStrLen])
	return blkHash, nil
}

// getSideEffectTransaction returns the side effect transaction for a block if said block has one.
// Side effects to reports are: genesis funds, undelegation payouts, permissioned-phase block rewards.
func (s *BlockAPI) getSideEffectTransaction(
	ctx context.Context, blk *itctypes.Block,
) (*types.Transaction, *types.Error) {
	if !s.containsSideEffectTransaction(ctx, blk) {
		return nil, common.NewError(common.TransactionNotFoundError, map[string]interface{}{
			"message": "no side effect transaction found for given block",
		})
	}

	var startingOpIndex *int64
	txOperations := []*types.Operation{}
	updateStartingOpIndex := func(newOperations []*types.Operation) {
		if len(newOperations) > 0 {
			index := newOperations[len(newOperations)-1].OperationIdentifier.Index + 1
			startingOpIndex = &index
		}
		txOperations = append(txOperations, newOperations...)
	}

	// Handle genesis funds
	if blk.NumberU64() == 0 {
		ops, rosettaError := GetSideEffectOperationsFromGenesisSpec(core.GetGenesisSpec(s.itc.ShardID), startingOpIndex)
		if rosettaError != nil {
			return nil, rosettaError
		}
		updateStartingOpIndex(ops)
	}
	// Handle block rewards for epoch < staking epoch (permissioned-phase block rewards)
	// Note that block rewards don't start until the second block.
	if !s.itc.IsStakingEpoch(blk.Epoch()) && blk.NumberU64() > 1 {
		rewards, err := s.itc.GetPreStakingBlockRewards(ctx, blk)
		if err != nil {
			return nil, common.NewError(common.CatchAllError, map[string]interface{}{
				"message": err.Error(),
			})
		}
		ops, rosettaError := GetSideEffectOperationsFromPreStakingRewards(rewards, startingOpIndex)
		if rosettaError != nil {
			return nil, rosettaError
		}
		updateStartingOpIndex(ops)
	}
	// Handle undelegation payout
	if s.itc.IsCommitteeSelectionBlock(blk.Header()) && s.itc.IsPreStakingEpoch(blk.Epoch()) {
		payouts, err := s.itc.GetUndelegationPayouts(ctx, blk.Epoch())
		if err != nil {
			return nil, common.NewError(common.CatchAllError, map[string]interface{}{
				"message": err.Error(),
			})
		}
		ops, rosettaError := getSideEffectOperationsFromUndelegationPayouts(payouts, startingOpIndex)
		if rosettaError != nil {
			return nil, rosettaError
		}
		updateStartingOpIndex(ops)
	}

	return &types.Transaction{
		TransactionIdentifier: getSideEffectTransactionIdentifier(blk.Hash()),
		Operations:            txOperations,
	}, nil
}

// sideEffectBlockTransaction is a formatter for side effect transactions
func (s *BlockAPI) sideEffectBlockTransaction(
	ctx context.Context, request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	// If no transaction info is found, check for special case transactions.
	blk, rosettaError := getBlock(ctx, s.itc, &types.PartialBlockIdentifier{Index: &request.BlockIdentifier.Index})
	if rosettaError != nil {
		return nil, rosettaError
	}
	blkHash, rosettaError := unpackSideEffectTransactionIdentifier(
		request.TransactionIdentifier,
	)
	if rosettaError != nil {
		return nil, rosettaError
	}
	if blkHash.String() != blk.Hash().String() {
		return nil, common.NewError(common.TransactionNotFoundError, map[string]interface{}{
			"message": fmt.Sprintf("side effect transaction is not for block: %v", blk.NumberU64()),
		})
	}
	tx, rosettaError := s.getSideEffectTransaction(ctx, blk)
	if rosettaError != nil {
		return nil, rosettaError
	}
	return &types.BlockTransactionResponse{Transaction: tx}, nil
}
