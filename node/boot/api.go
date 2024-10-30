package bootnode

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/zennittians/intelchain/eth/rpc"
	bootnodeConfigs "github.com/zennittians/intelchain/internal/configs/bootnode"
	nodeConfigs "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/internal/params"
	"github.com/zennittians/intelchain/internal/utils"
	itc_boot "github.com/zennittians/intelchain/itc_boot"
	boot_rpc "github.com/zennittians/intelchain/rpc/boot"
	rpc_common "github.com/zennittians/intelchain/rpc/boot/common"
)

// PeerID returns self Peer ID
func (bootnode *BootNode) PeerID() peer.ID {
	return bootnode.host.GetID()
}

// PeerConnectivity ..
func (bootnode *BootNode) PeerConnectivity() (int, int, int) {
	return bootnode.host.PeerConnectivity()
}

// ListKnownPeers return known peers
func (bootnode *BootNode) ListKnownPeers() peer.IDSlice {
	bs := bootnode.host.GetP2PHost().Peerstore()
	if bs == nil {
		return peer.IDSlice{}
	}
	return bs.Peers()
}

// ListConnectedPeers return connected peers
func (bootnode *BootNode) ListConnectedPeers() []peer.ID {
	return bootnode.host.Network().Peers()
}

// ListPeer return list of peers for a certain topic
func (bootnode *BootNode) ListPeer(topic string) []peer.ID {
	return bootnode.host.ListPeer(topic)
}

// ListTopic return list of topics the node subscribed
func (bootnode *BootNode) ListTopic() []string {
	return bootnode.host.ListTopic()
}

// ListBlockedPeer return list of blocked peers
func (bootnode *BootNode) ListBlockedPeer() []peer.ID {
	return bootnode.host.ListBlockedPeer()
}

// GetNodeBootTime ..
func (bootnode *BootNode) GetNodeBootTime() int64 {
	return bootnode.unixTimeAtNodeStart
}

// StartRPC start RPC service
func (bootnode *BootNode) StartRPC() error {
	bootService := itc_boot.New(bootnode)
	// Gather all the possible APIs to surface
	apis := bootnode.APIs(bootService)

	err := boot_rpc.StartServers(bootService, apis, *bootnode.RPCConfig, bootnode.IntelchainConfig.RPCOpt)

	return err
}

func (bootnode *BootNode) initRPCServerConfig() {
	cfg := bootnode.IntelchainConfig

	readTimeout, err := time.ParseDuration(cfg.HTTP.ReadTimeout)
	if err != nil {
		readTimeout, _ = time.ParseDuration(nodeConfigs.DefaultHTTPTimeoutRead)
		utils.Logger().Warn().
			Str("provided", cfg.HTTP.ReadTimeout).
			Dur("updated", readTimeout).
			Msg("Sanitizing invalid http read timeout")
	}
	writeTimeout, err := time.ParseDuration(cfg.HTTP.WriteTimeout)
	if err != nil {
		writeTimeout, _ = time.ParseDuration(nodeConfigs.DefaultHTTPTimeoutWrite)
		utils.Logger().Warn().
			Str("provided", cfg.HTTP.WriteTimeout).
			Dur("updated", writeTimeout).
			Msg("Sanitizing invalid http write timeout")
	}
	idleTimeout, err := time.ParseDuration(cfg.HTTP.IdleTimeout)
	if err != nil {
		idleTimeout, _ = time.ParseDuration(nodeConfigs.DefaultHTTPTimeoutIdle)
		utils.Logger().Warn().
			Str("provided", cfg.HTTP.IdleTimeout).
			Dur("updated", idleTimeout).
			Msg("Sanitizing invalid http idle timeout")
	}
	bootnode.RPCConfig = &bootnodeConfigs.RPCServerConfig{
		HTTPEnabled:        cfg.HTTP.Enabled,
		HTTPIp:             cfg.HTTP.IP,
		HTTPPort:           cfg.HTTP.Port,
		HTTPTimeoutRead:    readTimeout,
		HTTPTimeoutWrite:   writeTimeout,
		HTTPTimeoutIdle:    idleTimeout,
		WSEnabled:          cfg.WS.Enabled,
		WSIp:               cfg.WS.IP,
		WSPort:             cfg.WS.Port,
		DebugEnabled:       cfg.RPCOpt.DebugEnabled,
		RpcFilterFile:      cfg.RPCOpt.RpcFilterFile,
		RateLimiterEnabled: cfg.RPCOpt.RateLimterEnabled,
		RequestsPerSecond:  cfg.RPCOpt.RequestsPerSecond,
	}
}

func (bootnode *BootNode) GetRPCServerConfig() *bootnodeConfigs.RPCServerConfig {
	return bootnode.RPCConfig
}

// StopRPC stop RPC service
func (bootnode *BootNode) StopRPC() error {
	return boot_rpc.StopServers()
}

// APIs return the collection of local RPC services.
// NOTE, some of these services probably need to be moved to somewhere else.
func (bootnode *BootNode) APIs(intelchain *itc_boot.BootService) []rpc.API {
	// Append all the local APIs and return
	return []rpc.API{}
}

func (bootnode *BootNode) GetConfig() rpc_common.Config {
	return rpc_common.Config{
		IntelchainConfig: *bootnode.IntelchainConfig,
		NodeConfig:       *bootnode.NodeConfig,
		ChainConfig:      params.ChainConfig{},
	}
}
