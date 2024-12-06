package rpc

import (
	"context"

	"github.com/intelchain-itc/intelchain/core"
	"github.com/intelchain-itc/intelchain/eth/rpc"
	"github.com/intelchain-itc/intelchain/itc"
)

type PreimagesService struct {
	itc *itc.Intelchain
}

// NewPreimagesAPI creates a new API for the RPC interface
func NewPreimagesAPI(itc *itc.Intelchain, version string) rpc.API {
	var service interface{} = &PreimagesService{itc}
	return rpc.API{
		Namespace: version,
		Version:   APIVersion,
		Service:   service,
		Public:    true,
	}
}

func (s *PreimagesService) Export(ctx context.Context, path string) error {
	// these are by default not blocking
	return core.ExportPreimages(s.itc.BlockChain, path)
}

func (s *PreimagesService) Verify(ctx context.Context) (uint64, error) {
	currentBlock := s.itc.CurrentBlock()
	// these are by default not blocking
	return core.VerifyPreimages(currentBlock.Header(), s.itc.BlockChain)
}
