package itc_boot

import (
	commonRPC "github.com/zennittians/intelchain/rpc/boot/common"
)

// GetPeerInfo returns the peer info to the node, including blocked peer, connected peer, number of peers
func (itcboot *BootService) GetPeerInfo() commonRPC.BootNodePeerInfo {

	var c commonRPC.C
	c.TotalKnownPeers, c.Connected, c.NotConnected = itcboot.BootNodeAPI.PeerConnectivity()

	knownPeers := itcboot.BootNodeAPI.ListKnownPeers()
	connectedPeers := itcboot.BootNodeAPI.ListConnectedPeers()

	return commonRPC.BootNodePeerInfo{
		PeerID:         itcboot.BootNodeAPI.PeerID(),
		KnownPeers:     knownPeers,
		ConnectedPeers: connectedPeers,
		C:              c,
	}
}
