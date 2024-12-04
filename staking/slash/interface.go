package slash

import (
	"math/big"

	"github.com/intelchain-itc/intelchain/core/types"
	"github.com/intelchain-itc/intelchain/internal/params"
	"github.com/intelchain-itc/intelchain/shard"
)

// CommitteeReader ..
type CommitteeReader interface {
	Config() *params.ChainConfig
	ReadShardState(epoch *big.Int) (*shard.State, error)
	CurrentBlock() *types.Block
}
