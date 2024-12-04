package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/intelchain-itc/intelchain/eth/rpc"
	itcboot "github.com/intelchain-itc/intelchain/itc_boot"
)

// PublicBootService provides an API to access intelchain related information.
// It offers only methods that operate on public data that is freely available to anyone.
type PublicBootService struct {
	itcboot *itcboot.BootService
	version Version
}

// NewPublicBootAPI creates a new API for the RPC interface
func NewPublicBootAPI(itcboot *itcboot.BootService, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PublicBootService{itcboot, version},
		Public:    true,
	}
}

// ProtocolVersion returns the current intelchain protocol version this node supports
// Note that the return type is an interface to account for the different versions
func (s *PublicBootService) ProtocolVersion(
	ctx context.Context,
) (interface{}, error) {
	// Format response according to version
	switch s.version {
	case V1, Eth:
		return hexutil.Uint(s.itcboot.ProtocolVersion()), nil
	case V2:
		return s.itcboot.ProtocolVersion(), nil
	default:
		return nil, ErrUnknownRPCVersion
	}
}

// GetNodeMetadata produces a NodeMetadata record, data is from the answering RPC node
func (s *PublicBootService) GetNodeMetadata(
	ctx context.Context,
) (StructuredResponse, error) {
	// Response output is the same for all versions
	return NewStructuredResponse(s.itcboot.GetNodeMetadata())
}

// GetPeerInfo produces a NodePeerInfo record
func (s *PublicBootService) GetPeerInfo(
	ctx context.Context,
) (StructuredResponse, error) {
	// Response output is the same for all versions
	return NewStructuredResponse(s.itcboot.GetPeerInfo())
}
