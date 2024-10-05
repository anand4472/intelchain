package slash

import (
	"math/big"

	"github.com/zennittians/intelchain/core/types"
	"github.com/zennittians/intelchain/internal/params"
	"github.com/zennittians/intelchain/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
