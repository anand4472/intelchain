package node

import (
	"github.com/intelchain-itc/intelchain/consensus/quorum"
	"github.com/intelchain-itc/intelchain/core/types"
	"github.com/intelchain-itc/intelchain/eth/rpc"
	"github.com/intelchain-itc/intelchain/internal/tikv"
	"github.com/intelchain-itc/intelchain/itc"
	"github.com/intelchain-itc/intelchain/rosetta"
	itc_rpc "github.com/intelchain-itc/intelchain/rpc"
	rpc_common "github.com/intelchain-itc/intelchain/rpc/common"
	"github.com/intelchain-itc/intelchain/rpc/filters"
	"github.com/libp2p/go-libp2p/core/peer"
)

// IsCurrentlyLeader exposes if node is currently the leader node
func (node *Node) IsCurrentlyLeader() bool {
	return node.Consensus.IsLeader()
}

// PeerConnectivity ..
func (node *Node) PeerConnectivity() (int, int, int) {
	return node.host.PeerConnectivity()
}

// ListPeer return list of peers for a certain topic
func (node *Node) ListPeer(topic string) []peer.ID {
	return node.host.ListPeer(topic)
}

// ListTopic return list of topics the node subscribed
func (node *Node) ListTopic() []string {
	return node.host.ListTopic()
}

// ListBlockedPeer return list of blocked peers
func (node *Node) ListBlockedPeer() []peer.ID {
	return node.host.ListBlockedPeer()
}

// PendingCXReceipts returns node.pendingCXReceiptsProof
func (node *Node) PendingCXReceipts() []*types.CXReceiptsProof {
	cxReceipts := make([]*types.CXReceiptsProof, len(node.pendingCXReceipts))
	i := 0
	for _, cxReceipt := range node.pendingCXReceipts {
		cxReceipts[i] = cxReceipt
		i++
	}
	return cxReceipts
}

// ReportStakingErrorSink is the report of failed staking transactions this node has (held in memory only)
func (node *Node) ReportStakingErrorSink() types.TransactionErrorReports {
	return node.TransactionErrorSink.StakingReport()
}

// GetNodeBootTime ..
func (node *Node) GetNodeBootTime() int64 {
	return node.unixTimeAtNodeStart
}

// ReportPlainErrorSink is the report of failed transactions this node has (held in memory only)
func (node *Node) ReportPlainErrorSink() types.TransactionErrorReports {
	return node.TransactionErrorSink.PlainReport()
}

// StartRPC start RPC service
func (node *Node) StartRPC() error {
	intelchain := itc.New(node, node.TxPool, node.CxPool, node.Consensus.ShardID)

	// Gather all the possible APIs to surface
	apis := node.APIs(intelchain)

	return itc_rpc.StartServers(intelchain, apis, node.NodeConfig.RPCServer, node.IntelchainConfig.RPCOpt)
}

// StopRPC stop RPC service
func (node *Node) StopRPC() error {
	return itc_rpc.StopServers()
}

// StartRosetta start rosetta service
func (node *Node) StartRosetta() error {
	intelchain := itc.New(node, node.TxPool, node.CxPool, node.Consensus.ShardID)
	return rosetta.StartServers(intelchain, node.NodeConfig.RosettaServer, node.NodeConfig.RPCServer.RateLimiterEnabled, node.NodeConfig.RPCServer.RequestsPerSecond)
}

// StopRosetta stops rosetta service
func (node *Node) StopRosetta() error {
	return rosetta.StopServers()
}

// APIs return the collection of local RPC services.
// NOTE, some of these services probably need to be moved to somewhere else.
func (node *Node) APIs(intelchain *itc.Intelchain) []rpc.API {
	itcFilter := filters.NewPublicFilterAPI(intelchain, false, "itc", intelchain.ShardID)
	ethFilter := filters.NewPublicFilterAPI(intelchain, false, "eth", intelchain.ShardID)

	if node.IntelchainConfig.General.RunElasticMode && node.IntelchainConfig.TiKV.Role == tikv.RoleReader {
		itcFilter.Service.(*filters.PublicFilterAPI).SyncNewFilterFromOtherReaders()
		ethFilter.Service.(*filters.PublicFilterAPI).SyncNewFilterFromOtherReaders()
	}

	// Append all the local APIs and return
	return []rpc.API{
		itc_rpc.NewPublicNetAPI(node.host, intelchain.ChainID, itc_rpc.V1),
		itc_rpc.NewPublicNetAPI(node.host, intelchain.ChainID, itc_rpc.V2),
		itc_rpc.NewPublicNetAPI(node.host, intelchain.ChainID, itc_rpc.Eth),
		itc_rpc.NewPublicWeb3API(),
		itcFilter,
		ethFilter,
	}
}

// GetConsensusMode returns the current consensus mode
func (node *Node) GetConsensusMode() string {
	return node.Consensus.GetConsensusMode()
}

// GetConsensusPhase returns the current consensus phase
func (node *Node) GetConsensusPhase() string {
	return node.Consensus.GetConsensusPhase()
}

// GetConsensusViewChangingID returns the view changing ID
func (node *Node) GetConsensusViewChangingID() uint64 {
	return node.Consensus.GetViewChangingID()
}

// GetConsensusCurViewID returns the current view ID
func (node *Node) GetConsensusCurViewID() uint64 {
	return node.Consensus.GetCurBlockViewID()
}

// GetConsensusBlockNum returns the current block number of the consensus
func (node *Node) GetConsensusBlockNum() uint64 {
	return node.Consensus.BlockNum()
}

// GetConsensusInternal returns consensus internal data
func (node *Node) GetConsensusInternal() rpc_common.ConsensusInternal {
	return rpc_common.ConsensusInternal{
		ViewID:        node.GetConsensusCurViewID(),
		ViewChangeID:  node.GetConsensusViewChangingID(),
		Mode:          node.GetConsensusMode(),
		Phase:         node.GetConsensusPhase(),
		BlockNum:      node.GetConsensusBlockNum(),
		ConsensusTime: node.Consensus.GetFinality(),
	}
}

// IsBackup returns the node is in backup mode
func (node *Node) IsBackup() bool {
	return node.Consensus.IsBackup()
}

// SetNodeBackupMode change node backup mode
func (node *Node) SetNodeBackupMode(isBackup bool) bool {
	if node.Consensus.IsBackup() == isBackup {
		return false
	}

	node.Consensus.SetIsBackup(isBackup)
	node.Consensus.ResetViewChangeState()
	return true
}

func (node *Node) GetConfig() rpc_common.Config {
	return rpc_common.Config{
		IntelchainConfig: *node.IntelchainConfig,
		NodeConfig:       *node.NodeConfig,
		ChainConfig:      node.chainConfig,
	}
}

// GetLastSigningPower get last signed power
func (node *Node) GetLastSigningPower() (float64, error) {
	power, err := node.Consensus.Decider.CurrentTotalPower(quorum.Commit)
	if err != nil {
		return 0, err
	}

	round := float64(power.MulInt64(10000).RoundInt64()) / 10000
	return round, nil
}
