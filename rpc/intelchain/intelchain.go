package rpc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/zennittians/intelchain/eth/rpc"
	"github.com/zennittians/intelchain/itc"
)

// PublicIntelchainService provides an API to access Intelchain related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicIntelchainService struct {
	itc     *itc.Intelchain
	version Version
}

// NewPublicIntelchainAPI creates a new API for the RPC interface
func NewPublicIntelchainAPI(itc *itc.Intelchain, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PublicIntelchainService{itc, version},
		Public:    true,
	}
}

// ProtocolVersion returns the current Intelchain protocol version this node supports
// Note that the return type is an interface to account for the different versions
func (s *PublicIntelchainService) ProtocolVersion(
	ctx context.Context,
) (interface{}, error) {
	// Format response according to version
	switch s.version {
	case V1, Eth:
		return hexutil.Uint(s.itc.ProtocolVersion()), nil
	case V2:
		return s.itc.ProtocolVersion(), nil
	default:
		return nil, ErrUnknownRPCVersion
	}
}

// Syncing returns false in case the node is in sync with the network
// If it is syncing, it returns:
// starting block, current block, and network height
func (s *PublicIntelchainService) Syncing(
	ctx context.Context,
) (interface{}, error) {
	// difference = target - current
	inSync, target, difference := s.itc.NodeAPI.SyncStatus(s.itc.ShardID)
	if inSync {
		return false, nil
	}
	return struct {
		Start   uint64 `json:"startingBlock"`
		Current uint64 `json:"currentBlock"`
		Target  uint64 `json:"highestBlock"`
	}{
		// Start:   0, // TODO
		Current: target - difference,
		Target:  target,
	}, nil
}

// GasPrice returns a suggestion for a gas price.
// Note that the return type is an interface to account for the different versions
func (s *PublicIntelchainService) GasPrice(ctx context.Context) (interface{}, error) {
	price, err := s.itc.SuggestPrice(ctx)
	if err != nil || price.Cmp(big.NewInt(100e9)) < 0 {
		price = big.NewInt(100e9)
	}
	// Format response according to version
	switch s.version {
	case V1, Eth:
		return (*hexutil.Big)(price), nil
	case V2:
		return price.Uint64(), nil
	default:
		return nil, ErrUnknownRPCVersion
	}
}

// GetNodeMetadata produces a NodeMetadata record, data is from the answering RPC node
func (s *PublicIntelchainService) GetNodeMetadata(
	ctx context.Context,
) (StructuredResponse, error) {
	// Response output is the same for all versions
	return NewStructuredResponse(s.itc.GetNodeMetadata())
}

// GetPeerInfo produces a NodePeerInfo record
func (s *PublicIntelchainService) GetPeerInfo(
	ctx context.Context,
) (StructuredResponse, error) {
	// Response output is the same for all versions
	return NewStructuredResponse(s.itc.GetPeerInfo())
}

// GetNumPendingCrossLinks returns length of itc.BlockChain.ReadPendingCrossLinks()
func (s *PublicIntelchainService) GetNumPendingCrossLinks() (int, error) {
	links, err := s.itc.BlockChain.ReadPendingCrossLinks()
	if err != nil {
		return 0, err
	}

	return len(links), nil
}
