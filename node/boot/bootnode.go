package bootnode

import (
	"fmt"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
	"github.com/zennittians/intelchain/api/service"
	bootnodeConfigs "github.com/zennittians/intelchain/internal/configs/bootnode"
	intelchainConfig "github.com/zennittians/intelchain/internal/configs/intelchain"
	nodeConfig "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/p2p"
)

const (
	// NumTryBroadCast is the number of times trying to broadcast
	NumTryBroadCast = 3
	// MsgChanBuffer is the buffer of consensus message handlers.
	MsgChanBuffer = 1024
)

// BootNode represents a protocol-participating node in the network
type BootNode struct {
	SelfPeer p2p.Peer
	host     p2p.Host
	// Service manager.
	serviceManager *service.Manager
	// intelchain configurations
	IntelchainConfig *intelchainConfig.IntelchainConfig
	// node configuration, including group ID, shard ID, etc
	NodeConfig *nodeConfig.ConfigType
	// RPC configurations
	RPCConfig *bootnodeConfigs.RPCServerConfig
	// node start time
	unixTimeAtNodeStart int64
	// metrics
	Metrics metrics.Registry
}

// New creates a new boot node.
func New(
	host p2p.Host,
	hc *intelchainConfig.IntelchainConfig,
) *BootNode {
	node := BootNode{
		unixTimeAtNodeStart: time.Now().Unix(),
		IntelchainConfig:    hc,
		NodeConfig:          &nodeConfig.ConfigType{},
	}

	if host != nil {
		node.host = host
		node.SelfPeer = host.GetSelfPeer()
	}

	// init metrics
	initMetrics()
	nodeStringCounterVec.WithLabelValues("version", nodeConfig.GetVersion()).Inc()

	node.serviceManager = service.NewManager()

	node.initRPCServerConfig()

	return &node
}

// ServiceManager ...
func (bootnode *BootNode) ServiceManager() *service.Manager {
	return bootnode.serviceManager
}

// ShutDown gracefully shut down the node server and dump the in-memory blockchain state into DB.
func (bootnode *BootNode) ShutDown() {
	if err := bootnode.StopRPC(); err != nil {
		utils.Logger().Error().Err(err).Msg("failed to stop boot RPC")
	}

	utils.Logger().Info().Msg("stopping boot services")
	if err := bootnode.StopServices(); err != nil {
		utils.Logger().Error().Err(err).Msg("failed to stop boot services")
	}

	utils.Logger().Info().Msg("stopping boot host")
	if err := bootnode.host.Close(); err != nil {
		utils.Logger().Error().Err(err).Msg("failed to stop boot p2p host")
	}

	const msg = "Successfully shut down boot!\n"
	utils.Logger().Print(msg)
	fmt.Print(msg)
	os.Exit(0)
}
