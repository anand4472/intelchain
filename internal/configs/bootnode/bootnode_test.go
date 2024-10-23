package bootnode

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
)

func TestToRPCServerConfig(t *testing.T) {
	tests := []struct {
		input  BootNodeConfig
		output nodeconfig.RPCServerConfig
	}{
		{
			input: BootNodeConfig{
				HTTP: HttpConfig{
					Enabled:      true,
					IP:           "146.190.38.146",
					Port:         nodeconfig.DefaultRPCPort,
					ReadTimeout:  "-1",
					WriteTimeout: "-2",
					IdleTimeout:  "-3",
				},
				WS: WsConfig{
					Enabled: true,
					IP:      "146.190.38.146",
					Port:    nodeconfig.DefaultWSPort,
				},
				RPCOpt: RpcOptConfig{
					DebugEnabled:      false,
					EthRPCsEnabled:    true,
					LegacyRPCsEnabled: true,
					RpcFilterFile:     "./.itc/rpc_filter.txt",
					RateLimterEnabled: true,
					RequestsPerSecond: nodeconfig.DefaultRPCRateLimit,
				},
			},
			output: nodeconfig.RPCServerConfig{
				HTTPEnabled:        true,
				HTTPIp:             "146.190.38.146",
				HTTPPort:           nodeconfig.DefaultRPCPort,
				HTTPTimeoutRead:    30 * time.Second,
				HTTPTimeoutWrite:   30 * time.Second,
				HTTPTimeoutIdle:    120 * time.Second,
				WSEnabled:          true,
				WSIp:               "146.190.38.146",
				WSPort:             nodeconfig.DefaultWSPort,
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimiterEnabled: true,
				RequestsPerSecond:  nodeconfig.DefaultRPCRateLimit,
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
