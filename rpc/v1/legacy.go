package v1

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/intelchain-itc/intelchain/eth/rpc"
	internal_common "github.com/intelchain-itc/intelchain/internal/common"
	"github.com/intelchain-itc/intelchain/itc"
)

// PublicLegacyService provides an API to access the Intelchain blockchain.
// Services here are legacy methods, specific to the V1 RPC that can be deprecated in the future.
type PublicLegacyService struct {
	itc *itc.Intelchain
}

// NewPublicLegacyAPI creates a new API for the RPC interface
func NewPublicLegacyAPI(itc *itc.Intelchain, namespace string) rpc.API {
	if namespace == "" {
		namespace = "itc"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicLegacyService{itc},
		Public:    true,
	}
}

// GetBalance returns the amount of TICK for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicLegacyService) GetBalance(
	ctx context.Context, address string, blockNrOrHash rpc.BlockNumberOrHash,
) (*hexutil.Big, error) {
	addr, err := internal_common.ParseAddr(address)
	if err != nil {
		return nil, err
	}
	balance, err := s.itc.GetBalance(ctx, addr, blockNrOrHash)
	if err != nil {
		return nil, err
	}
	return (*hexutil.Big)(balance), nil
}
