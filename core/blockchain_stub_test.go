package core_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zennittians/intelchain/core"
)

type bc struct {
	core.Stub
}

func TestName(t *testing.T) {
	require.EqualError(t, bc{Stub: core.Stub{Name: "Core"}}.SetHead(0), "method SetHead not implemented for Core")
}
