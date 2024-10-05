package rpc

import (
	"context"

	"github.com/zennittians/intelchain/eth/rpc"
	"github.com/zennittians/intelchain/itc"
)

// PrivateDebugService Internal JSON RPC for debugging purpose
type PrivateDebugService struct {
	itc     *itc.Intelchain
	version Version
}

// NewPrivateDebugAPI creates a new API for the RPC interface
// TODO(dm): expose public via config
func NewPrivateDebugAPI(itc *itc.Intelchain, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PrivateDebugService{itc, version},
		Public:    false,
	}
}

// ConsensusViewChangingID return the current view changing ID to RPC
func (s *PrivateDebugService) ConsensusViewChangingID(
	ctx context.Context,
) uint64 {
	return s.itc.NodeAPI.GetConsensusViewChangingID()
}

// ConsensusCurViewID return the current view ID to RPC
func (s *PrivateDebugService) ConsensusCurViewID(
	ctx context.Context,
) uint64 {
	return s.itc.NodeAPI.GetConsensusCurViewID()
}

// GetConsensusMode return the current consensus mode
func (s *PrivateDebugService) GetConsensusMode(
	ctx context.Context,
) string {
	return s.itc.NodeAPI.GetConsensusMode()
}

// GetConsensusPhase return the current consensus mode
func (s *PrivateDebugService) GetConsensusPhase(
	ctx context.Context,
) string {
	return s.itc.NodeAPI.GetConsensusPhase()
}

// GetConfig get Intelchain config
func (s *PrivateDebugService) GetConfig(
	ctx context.Context,
) (StructuredResponse, error) {
	return NewStructuredResponse(s.itc.NodeAPI.GetConfig())
}

// GetLastSigningPower get last signed power
func (s *PrivateDebugService) GetLastSigningPower(
	ctx context.Context,
) (float64, error) {
	return s.itc.NodeAPI.GetLastSigningPower()
}

// GetLastSigningPower2 get last signed power
func (s *PrivateDebugService) GetLastSigningPower2(
	ctx context.Context,
) (float64, error) {
	return s.itc.NodeAPI.GetLastSigningPower2()
}
