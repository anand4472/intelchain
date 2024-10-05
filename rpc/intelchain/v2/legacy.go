package v2

import (
	"context"
	"math/big"

	"github.com/zennittians/intelchain/eth/rpc"
	internal_common "github.com/zennittians/intelchain/internal/common"
	"github.com/zennittians/intelchain/itc"
)

// PublicLegacyService provides an API to access the intelchain blockchain.
// Services here are legacy methods, specific to the V1 RPC that can be deprecated in the future.
type PublicLegacyService struct {
	itc *itc.Intelchain
}

// NewPublicLegacyAPI creates a new API for the RPC interface
func NewPublicLegacyAPI(itc *itc.Intelchain, namespace string) rpc.API {
	if namespace == "" {
		namespace = "itcv2"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicLegacyService{itc},
		Public:    true,
	}
}

// GetBalance returns the amount of Intelli for the given address in the state of the
// given block number.
func (s *PublicLegacyService) GetBalance(
	ctx context.Context, address string,
) (*big.Int, error) {
	addr, err := internal_common.ParseAddr(address)
	if err != nil {
		return nil, err
	}
	balance, err := s.itc.GetBalance(ctx, addr, rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber))
	if err != nil {
		return nil, err
	}
	return balance, nil
}
