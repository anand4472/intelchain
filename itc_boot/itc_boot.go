package itc_boot

import (
	"github.com/intelchain-itc/intelchain/api/proto"
	nodeconfig "github.com/intelchain-itc/intelchain/internal/configs/node"
	rpc_common "github.com/intelchain-itc/intelchain/rpc/boot/common"
	"github.com/libp2p/go-libp2p/core/peer"
)

// BootService implements the BootService full node service.
type BootService struct {
	// Channel for shutting down the service
	ShutdownChan chan bool // Channel for shutting down the BootService
	// Boot node API
	BootNodeAPI BootNodeAPI
	// Shard ID
	ShardID uint32
}

// BootNodeAPI is the list of functions from node used to call rpc apis.
type BootNodeAPI interface {
	GetNodeBootTime() int64
	PeerID() peer.ID
	PeerConnectivity() (int, int, int)
	ListKnownPeers() peer.IDSlice
	ListConnectedPeers() []peer.ID
	ListPeer(topic string) []peer.ID
	ListTopic() []string
	ListBlockedPeer() []peer.ID
	GetConfig() rpc_common.Config
	ShutDown()
}

// New creates a new BootService object (including the
// initialisation of the common BootService object)
func New(nodeAPI BootNodeAPI) *BootService {
	backend := &BootService{
		ShutdownChan: make(chan bool),

		BootNodeAPI: nodeAPI,
	}

	return backend
}

// ProtocolVersion ...
func (itcboot *BootService) ProtocolVersion() int {
	return proto.ProtocolVersion
}

// GetNodeMetadata returns the node metadata.
func (itcboot *BootService) GetNodeMetadata() rpc_common.BootNodeMetadata {
	var c rpc_common.C

	c.TotalKnownPeers, c.Connected, c.NotConnected = itcboot.BootNodeAPI.PeerConnectivity()

	return rpc_common.BootNodeMetadata{
		Version:      nodeconfig.GetVersion(),
		Network:      string(nodeconfig.GetDefaultConfig().GetNetworkType()),
		NodeBootTime: itcboot.BootNodeAPI.GetNodeBootTime(),
		PeerID:       nodeconfig.GetPeerID(),
		C:            c,
	}
}
