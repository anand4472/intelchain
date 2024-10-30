package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zennittians/intelchain/core"
)

func TestRegistry(t *testing.T) {
	registry := New()
	require.Nil(t, registry.GetBlockchain())

	registry.SetBlockchain(core.Stub{})
	require.NotNil(t, registry.GetBlockchain())
}
