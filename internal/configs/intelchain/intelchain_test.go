package intelchain

import (
	"fmt"
	"testing"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
)

func TestToRPCServerConfig(t *testing.T) {
	tests := []struct {
		input  IntelchainConfig
		output nodeconfig.RPCServerConfig
	}{
		{
			input: IntelchainConfig{
				HTTP: HttpConfig{
					Enabled:        true,
					RosettaEnabled: false,
					IP:             "146.190.38.146",
					Port:           nodeconfig.DefaultRPCPort,
					AuthPort:       nodeconfig.DefaultAuthRPCPort,
					RosettaPort:    nodeconfig.DefaultRosettaPort,
					ReadTimeout:    "-1",
					WriteTimeout:   "-2",
					IdleTimeout:    "-3",
				},
				WS: WsConfig{
					Enabled:  true,
					IP:       "146.190.38.146",
					Port:     nodeconfig.DefaultWSPort,
					AuthPort: nodeconfig.DefaultAuthWSPort,
				},
				RPCOpt: RpcOptConfig{
					DebugEnabled:       false,
					EthRPCsEnabled:     true,
					StakingRPCsEnabled: true,
					LegacyRPCsEnabled:  true,
					RpcFilterFile:      "./.itc/rpc_filter.txt",
					RateLimterEnabled:  true,
					RequestsPerSecond:  nodeconfig.DefaultRPCRateLimit,
					EvmCallTimeout:     "-4",
				},
			},
			output: nodeconfig.RPCServerConfig{
				HTTPEnabled:        true,
				HTTPIp:             "146.190.38.146",
				HTTPPort:           nodeconfig.DefaultRPCPort,
				HTTPAuthPort:       nodeconfig.DefaultAuthRPCPort,
				HTTPTimeoutRead:    30 * time.Second,
				HTTPTimeoutWrite:   30 * time.Second,
				HTTPTimeoutIdle:    120 * time.Second,
				WSEnabled:          true,
				WSIp:               "146.190.38.146",
				WSPort:             nodeconfig.DefaultWSPort,
				WSAuthPort:         nodeconfig.DefaultAuthWSPort,
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimiterEnabled: true,
				RequestsPerSecond:  nodeconfig.DefaultRPCRateLimit,
				EvmCallTimeout:     5 * time.Second,
			},
		},
	}
	for i, tt := range tests {
		assertObject := assert.New(t)
		name := fmt.Sprintf("TestToRPCServerConfig: #%d", i)
		t.Run(name, func(t *testing.T) {
			assertObject.Equal(
				tt.input.ToRPCServerConfig(),
				tt.output,
				name,
			)
		})
	}
}

var data = `big = 100e9
small = 100
zero = 0
`

func TestPriceLimit_UnmarshalTOML(t *testing.T) {
	type V struct {
		Big   PriceLimit `toml:"big"`
		Small PriceLimit `toml:"small"`
		Zero  PriceLimit `toml:"zero"`
	}
	var v V
	require.NoError(t, toml.Unmarshal([]byte(data), &v))

	require.Equal(t, PriceLimit(100e9), v.Big)
	require.Equal(t, PriceLimit(100), v.Small)
	require.Equal(t, PriceLimit(0), v.Zero)
}

func TestPriceLimit_MarshalTOML(t *testing.T) {
	type V struct {
		Big   PriceLimit `toml:"big"`
		Small PriceLimit `toml:"small"`
		Zero  PriceLimit `toml:"zero"`
	}
	v := V{
		Big:   PriceLimit(100e9),
		Small: PriceLimit(100),
		Zero:  PriceLimit(0),
	}
	e, err := toml.Marshal(v)
	require.NoError(t, err)
	require.Equal(t, data, string(e))
}
