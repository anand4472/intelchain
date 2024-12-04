package registry

import (
	"testing"

	"github.com/intelchain-itc/intelchain/core"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	registry := New()
	require.Nil(t, registry.GetBlockchain())

	registry.SetBlockchain(core.Stub{})
	require.NotNil(t, registry.GetBlockchain())
}
