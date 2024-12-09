package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/intelchain-itc/intelchain/eth/rpc"
	internal_common "github.com/intelchain-itc/intelchain/internal/common"
	"github.com/intelchain-itc/intelchain/itc"
)

// PublicEthService provides an API to access to the Eth endpoints for the intelchain blockchain.
type PublicEthService struct {
	itc *itc.Intelchain
}

// NewPublicEthService creates a new API for the RPC interface
func NewPublicEthService(itc *itc.Intelchain, namespace string) rpc.API {
	if namespace == "" {
		namespace = "eth"
	}

	return rpc.API{
		Namespace: namespace,
		Version:   "1.0",
		Service:   &PublicEthService{itc},
		Public:    true,
	}
}

// GetBalance returns the amount of Atto for the given address in the state of the
// given block number. The rpc.LatestBlockNumber and rpc.PendingBlockNumber meta
// block numbers are also allowed.
func (s *PublicEthService) GetBalance(
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
