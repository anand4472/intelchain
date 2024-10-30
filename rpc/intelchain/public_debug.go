package rpc

import (
	"context"

	"github.com/ethereum/go-ethereum/log"
	"github.com/zennittians/intelchain/eth/rpc"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/itc"
)

// PublicDebugService Internal JSON RPC for debugging purpose
type PublicDebugService struct {
	itc     *itc.Intelchain
	version Version
}

// NewPublicDebugAPI creates a new API for the RPC interface
func NewPublicDebugAPI(itc *itc.Intelchain, version Version) rpc.API {
	return rpc.API{
		Namespace: version.Namespace(),
		Version:   APIVersion,
		Service:   &PublicDebugService{itc, version},
		Public:    false,
	}
}

// SetLogVerbosity Sets log verbosity on runtime
// curl -H "Content-Type: application/json" -d '{"method":"itc_setLogVerbosity","params":[5],"id":1}' http://127.0.0.1:9500
func (s *PublicDebugService) SetLogVerbosity(ctx context.Context, level int) (map[string]interface{}, error) {
	if level < int(log.LvlCrit) || level > int(log.LvlTrace) {
		return nil, ErrInvalidLogLevel
	}

	verbosity := log.Lvl(level)
	utils.SetLogVerbosity(verbosity)
	return map[string]interface{}{"verbosity": verbosity.String()}, nil
}
