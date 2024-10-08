package core_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zennittians/intelchain/core"
	"github.com/zennittians/intelchain/core/rawdb"
	"github.com/zennittians/intelchain/core/vm"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
)

func TestGenesisBlock(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	err := (&core.GenesisInitializer{NetworkType: nodeconfig.Mainnet}).InitChainDB(db, 0)
	require.NoError(t, err)

	chain, err := core.NewEpochChain(db, nil, nil, vm.Config{})
	require.NoError(t, err)

	header := chain.GetHeaderByNumber(0)
	require.NotEmpty(t, header)
}
