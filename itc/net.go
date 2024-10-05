package itc

import (
	"github.com/libp2p/go-libp2p/core/peer"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	commonRPC "github.com/zennittians/intelchain/rpc/intelchain/common"
	"github.com/zennittians/intelchain/staking/network"
)

// GetCurrentUtilityMetrics ..
func (itc *Intelchain) GetCurrentUtilityMetrics() (*network.UtilityMetric, error) {
	return network.NewUtilityMetricSnapshot(itc.BlockChain)
}

// GetPeerInfo returns the peer info to the node, including blocked peer, connected peer, number of peers
func (itc *Intelchain) GetPeerInfo() commonRPC.NodePeerInfo {

	topics := itc.NodeAPI.ListTopic()
	p := make([]commonRPC.P, len(topics))

	for i, t := range topics {
		topicPeer := itc.NodeAPI.ListPeer(t)
		p[i].Topic = t
		p[i].Peers = make([]peer.ID, len(topicPeer))
		copy(p[i].Peers, topicPeer)
	}

	return commonRPC.NodePeerInfo{
		PeerID:       nodeconfig.GetPeerID(),
		BlockedPeers: itc.NodeAPI.ListBlockedPeer(),
		P:            p,
	}
}
