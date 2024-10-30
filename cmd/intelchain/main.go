package main

import (
	"fmt"
	"math/big"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zennittians/bls/ffi/go/bls"
	"github.com/zennittians/intelchain/api/service"
	"github.com/zennittians/intelchain/api/service/crosslink_sending"
	"github.com/zennittians/intelchain/api/service/pprof"
	"github.com/zennittians/intelchain/api/service/prometheus"
	"github.com/zennittians/intelchain/api/service/stagedstreamsync"
	"github.com/zennittians/intelchain/api/service/synchronize"
	intelchainConfigs "github.com/zennittians/intelchain/cmd/config"
	"github.com/zennittians/intelchain/common/fdlimit"
	"github.com/zennittians/intelchain/common/ntp"
	"github.com/zennittians/intelchain/consensus"
	"github.com/zennittians/intelchain/consensus/quorum"
	"github.com/zennittians/intelchain/core"
	"github.com/zennittians/intelchain/internal/chain"
	"github.com/zennittians/intelchain/internal/cli"
	"github.com/zennittians/intelchain/internal/common"
	intelchainconfig "github.com/zennittians/intelchain/internal/configs/intelchain"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	shardingconfig "github.com/zennittians/intelchain/internal/configs/sharding"
	"github.com/zennittians/intelchain/internal/genesis"
	"github.com/zennittians/intelchain/internal/params"
	"github.com/zennittians/intelchain/internal/registry"
	"github.com/zennittians/intelchain/internal/shardchain"
	"github.com/zennittians/intelchain/internal/shardchain/tikv_manage"
	"github.com/zennittians/intelchain/internal/tikv/redis_helper"
	"github.com/zennittians/intelchain/internal/tikv/statedb_cache"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/itc/downloader"
	"github.com/zennittians/intelchain/multibls"
	node "github.com/zennittians/intelchain/node/intelchain"
	"github.com/zennittians/intelchain/numeric"
	"github.com/zennittians/intelchain/p2p"
	rosetta_common "github.com/zennittians/intelchain/rosetta/common"
	rpc_common "github.com/zennittians/intelchain/rpc/intelchain/common"
	"github.com/zennittians/intelchain/shard"
	"github.com/zennittians/intelchain/webhooks"
)

// Version string variables
var (
	version string
	builtBy string
	builtAt string
	commit  string
)

// Host
var (
	myHost          p2p.Host
	initialAccounts = []*genesis.DeployAccount{}
)

var rootCmd = &cobra.Command{
	Use:   "intelchain",
	Short: "intelchain is the intelchain node binary file",
	Long: `intelchain is the intelchain node binary file

Examples usage:

# start a validator node with default bls folder (default bls key files in ./.itc/blskeys)
    ./intelchain

# start a validator node with customized bls key folder
    ./intelchain --bls.dir [bls_folder]

# start a validator node with open RPC endpoints and customized ports
    ./intelchain --http.ip=0.0.0.0 --http.port=[http_port] --ws.ip=0.0.0.0 --ws.port=[ws_port]

# start an explorer node
    ./intelchain --run=explorer --run.shard=[shard_id]

# start a intelchain internal node on testnet
    ./intelchain --run.legacy --network testnet
`,
	Run: runIntelchainNode,
}

func init() {
	intelchainConfigs.VersionMetaData = append(intelchainConfigs.VersionMetaData, "intelchain", version, commit, builtBy, builtAt)
	intelchainConfigs.Init(rootCmd)
}

func main() {
	rootCmd.Execute()
}

func runIntelchainNode(cmd *cobra.Command, args []string) {
	if cli.GetBoolFlagValue(cmd, intelchainConfigs.VersionFlag()) {
		intelchainConfigs.PrintVersion()
		os.Exit(0)
	}

	if err := prepareRootCmd(cmd); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(128)
	}
	cfg, err := intelchainConfigs.GetIntelchainConfig(cmd)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(128)
	}

	setupNodeLog(cfg)
	setupNodeAndRun(cfg)
}

func prepareRootCmd(cmd *cobra.Command) error {
	// HACK Force usage of go implementation rather than the C based one. Do the right way, see the
	// notes one line 66,67 of https://golang.org/src/net/net.go that say can make the decision at
	// build time.
	os.Setenv("GODEBUG", "netdns=go")
	// Don't set higher than num of CPU. It will make go scheduler slower.
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Raise fd limits
	return raiseFdLimits()
}

func raiseFdLimits() error {
	limit, err := fdlimit.Maximum()
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve file descriptor allowance")
	}
	_, err = fdlimit.Raise(uint64(limit))
	if err != nil {
		return errors.Wrap(err, "Failed to raise file descriptor allowance")
	}
	return nil
}

func setupNodeLog(config intelchainconfig.IntelchainConfig) {
	logPath := filepath.Join(config.Log.Folder, config.Log.FileName)
	verbosity := config.Log.Verbosity

	utils.SetLogVerbosity(log.Lvl(verbosity))
	if config.Log.Context != nil {
		ip := config.Log.Context.IP
		port := config.Log.Context.Port
		utils.SetLogContext(ip, strconv.Itoa(port))
	}

	if !config.Log.Console {
		utils.AddLogFile(logPath, config.Log.RotateSize, config.Log.RotateCount, config.Log.RotateMaxAge)
	}
}

func revert(chain core.BlockChain, hc intelchainconfig.IntelchainConfig) {
	curNum := chain.CurrentBlock().NumberU64()
	if curNum < uint64(hc.Revert.RevertBefore) && curNum >= uint64(hc.Revert.RevertTo) {
		// Remove invalid blocks
		for chain.CurrentBlock().NumberU64() >= uint64(hc.Revert.RevertTo) {
			curBlock := chain.CurrentBlock()
			rollbacks := []ethCommon.Hash{curBlock.Hash()}
			if err := chain.Rollback(rollbacks); err != nil {
				fmt.Printf("Revert failed: %v\n", err)
				os.Exit(1)
			}
			lastSig := curBlock.Header().LastCommitSignature()
			sigAndBitMap := append(lastSig[:], curBlock.Header().LastCommitBitmap()...)
			chain.WriteCommitSig(curBlock.NumberU64()-1, sigAndBitMap)
		}
		fmt.Printf("Revert finished. Current block: %v\n", chain.CurrentBlock().NumberU64())
		utils.Logger().Warn().
			Uint64("Current Block", chain.CurrentBlock().NumberU64()).
			Msg("Revert finished.")
		os.Exit(1)
	}
}

func setupNodeAndRun(hc intelchainconfig.IntelchainConfig) {
	var err error

	nodeconfigSetShardSchedule(hc)
	nodeconfig.SetShardingSchedule(shard.Schedule)
	nodeconfig.SetVersion(intelchainConfigs.GetIntelchainVersion())

	if hc.General.NodeType == "validator" {
		var err error
		if hc.General.NoStaking {
			err = setupLegacyNodeAccount(hc)
		} else {
			err = setupStakingNodeAccount(hc)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot set up node account: %s\n", err)
			os.Exit(1)
		}
	}
	if hc.General.NodeType == "validator" {
		fmt.Printf("%s mode; node key %s -> shard %d\n",
			map[bool]string{false: "Legacy", true: "Staking"}[!hc.General.NoStaking],
			nodeconfig.GetDefaultConfig().ConsensusPriKey.GetPublicKeys().SerializeToHexStr(),
			initialAccounts[0].ShardID)
	}
	if hc.General.NodeType != "validator" && hc.General.ShardID >= 0 {
		for _, initialAccount := range initialAccounts {
			utils.Logger().Info().
				Uint32("original", initialAccount.ShardID).
				Int("override", hc.General.ShardID).
				Msg("ShardID Override")
			initialAccount.ShardID = uint32(hc.General.ShardID)
		}
	}

	nodeConfig, err := createGlobalConfig(hc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR cannot configure node: %s\n", err)
		os.Exit(1)
	}

	if hc.General.RunElasticMode && hc.TiKV == nil {
		fmt.Fprintf(os.Stderr, "Use TIKV MUST HAS TIKV CONFIG")
		os.Exit(1)
	}

	// Update ethereum compatible chain ids
	params.UpdateEthChainIDByShard(nodeConfig.ShardID)

	currentNode := setupConsensusAndNode(hc, nodeConfig, registry.New())
	nodeconfig.GetDefaultConfig().ShardID = nodeConfig.ShardID
	nodeconfig.GetDefaultConfig().IsOffline = nodeConfig.IsOffline
	nodeconfig.GetDefaultConfig().Downloader = nodeConfig.Downloader
	nodeconfig.GetDefaultConfig().StagedSync = nodeConfig.StagedSync

	// Check NTP and time accuracy
	// It skips the time accuracy check on the localnet since all nodes are running on the same machine
	if hc.Network.NetworkType != nodeconfig.Localnet {
		clockAccuracyResp, err := ntp.CheckLocalTimeAccurate(nodeConfig.NtpServer)
		if !clockAccuracyResp.IsAccurate() {
			if clockAccuracyResp.AllNtpServersTimedOut() {
				fmt.Fprintf(os.Stderr, "Error: querying NTP servers timed out, Continuing.\n")
			} else if clockAccuracyResp.NtpFailed() {
				fmt.Fprintf(os.Stderr, "Error: NTP servers are not properly configured, %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Error: local time clock is not accurate, %s\n", clockAccuracyResp.Message())
			}
		}
		if err != nil {
			utils.Logger().Warn().Err(err).Msg("Check Local Time Accuracy Error")
		}
	}

	// Parse RPC config
	nodeConfig.RPCServer = hc.ToRPCServerConfig()

	// Parse rosetta config
	nodeConfig.RosettaServer = nodeconfig.RosettaServerConfig{
		HTTPEnabled: hc.HTTP.RosettaEnabled,
		HTTPIp:      hc.HTTP.IP,
		HTTPPort:    hc.HTTP.RosettaPort,
	}

	if hc.Revert != nil && hc.Revert.RevertBefore != 0 && hc.Revert.RevertTo != 0 {
		chain := currentNode.Blockchain()
		if hc.Revert.RevertBeacon {
			chain = currentNode.Beaconchain()
		}
		revert(chain, hc)
	}

	//// code to handle pre-image export, import and generation
	if hc.Preimage != nil {
		if hc.Preimage.ImportFrom != "" {
			if err := core.ImportPreimages(
				currentNode.Blockchain(),
				hc.Preimage.ImportFrom,
			); err != nil {
				fmt.Println("Error importing", err)
				os.Exit(1)
			}
			os.Exit(0)
		} else if exportPath := hc.Preimage.ExportTo; exportPath != "" {
			if err := core.ExportPreimages(
				currentNode.Blockchain(),
				exportPath,
			); err != nil {
				fmt.Println("Error exporting", err)
				os.Exit(1)
			}
			os.Exit(0)
			// both must be set
		} else if hc.Preimage.GenerateStart > 0 {
			chain := currentNode.Blockchain()
			end := hc.Preimage.GenerateEnd
			current := chain.CurrentBlock().NumberU64()
			if end > current {
				fmt.Printf(
					"Cropping generate endpoint from %d to %d\n",
					end, current,
				)
				end = current
			}

			if end == 0 {
				end = current
			}

			fmt.Println("Starting generation")
			if err := core.GeneratePreimages(
				chain,
				hc.Preimage.GenerateStart, end,
			); err != nil {
				fmt.Println("Error generating", err)
				os.Exit(1)
			}
			fmt.Println("Generation successful")
			os.Exit(0)
		}
		os.Exit(0)
	}

	startMsg := "==== New Intelchain Node ===="
	if hc.General.NodeType == intelchainConfigs.NodeTypeExplorer {
		startMsg = "==== New Explorer Node ===="
	}

	utils.Logger().Info().
		Str("BLSPubKey", nodeConfig.ConsensusPriKey.GetPublicKeys().SerializeToHexStr()).
		Uint32("ShardID", nodeConfig.ShardID).
		Str("ShardGroupID", nodeConfig.GetShardGroupID().String()).
		Str("BeaconGroupID", nodeConfig.GetBeaconGroupID().String()).
		Str("ClientGroupID", nodeConfig.GetClientGroupID().String()).
		Str("Role", currentNode.NodeConfig.Role().String()).
		Str("Version", intelchainConfigs.GetIntelchainVersion()).
		Str("multiaddress",
			fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", hc.P2P.IP, hc.P2P.Port, myHost.GetID().String()),
		).
		Msg(startMsg)

	nodeconfig.SetPeerID(myHost.GetID())

	if hc.Log.VerbosePrints.Config {
		utils.Logger().Info().Interface("config", rpc_common.Config{
			IntelchainConfig: hc,
			NodeConfig:       *nodeConfig,
			ChainConfig:      *currentNode.Blockchain().Config(),
		}).Msg("verbose prints config")
	}

	// Setup services
	if hc.Sync.Enabled {
		if hc.Sync.StagedSync {
			setupStagedSyncService(currentNode, myHost, hc)
		} else {
			setupSyncService(currentNode, myHost, hc)
		}
	}
	if currentNode.NodeConfig.Role() == nodeconfig.Validator {
		currentNode.RegisterValidatorServices()
	} else if currentNode.NodeConfig.Role() == nodeconfig.ExplorerNode {
		currentNode.RegisterExplorerServices()
	}
	currentNode.RegisterService(service.CrosslinkSending, crosslink_sending.New(currentNode, currentNode.Blockchain()))
	if hc.Pprof.Enabled {
		setupPprofService(currentNode, hc)
	}
	if hc.Prometheus.Enabled {
		setupPrometheusService(currentNode, hc, nodeConfig.ShardID)
	}

	if hc.DNSSync.Server && !hc.General.IsOffline {
		utils.Logger().Info().Msg("support gRPC sync server")
		currentNode.SupportGRPCSyncServer(hc.DNSSync.ServerPort)
	}
	if hc.DNSSync.Client && !hc.General.IsOffline {
		utils.Logger().Info().Msg("go with gRPC sync client")
		currentNode.StartGRPCSyncClient()
	}

	currentNode.NodeSyncing()

	if err := currentNode.StartServices(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(-1)
	}

	if err := currentNode.StartRPC(); err != nil {
		utils.Logger().Warn().
			Err(err).
			Msg("StartRPC failed")
	}

	if err := currentNode.StartRosetta(); err != nil {
		utils.Logger().Warn().
			Err(err).
			Msg("Start Rosetta failed")
	}

	go core.WritePreimagesMetricsIntoPrometheus(
		currentNode.Blockchain(),
		currentNode.Consensus.UpdatePreimageGenerationMetrics,
	)

	go listenOSSigAndShutDown(currentNode)

	if !hc.General.IsOffline {
		if err := myHost.Start(); err != nil {
			utils.Logger().Fatal().
				Err(err).
				Msg("Start p2p host failed")
		}
		

		if err := currentNode.BootstrapConsensus(); err != nil {
			fmt.Fprint(os.Stderr, "could not bootstrap consensus", err.Error())
			if !currentNode.NodeConfig.IsOffline {
				os.Exit(-1)
			}
		}

		if err := currentNode.StartPubSub(); err != nil {
			fmt.Fprint(os.Stderr, "could not begin network message handling for node", err.Error())
			os.Exit(-1)
		}
	}

	select {}
}

func nodeconfigSetShardSchedule(config intelchainconfig.IntelchainConfig) {
	switch config.Network.NetworkType {
	case nodeconfig.Mainnet:
		shard.Schedule = shardingconfig.MainnetSchedule
	case nodeconfig.Testnet:
		shard.Schedule = shardingconfig.TestnetSchedule
	case nodeconfig.Pangaea:
		shard.Schedule = shardingconfig.PangaeaSchedule
	case nodeconfig.Localnet:
		shard.Schedule = shardingconfig.LocalnetSchedule
	case nodeconfig.Partner:
		shard.Schedule = shardingconfig.PartnerSchedule
	case nodeconfig.Stressnet:
		shard.Schedule = shardingconfig.StressNetSchedule
	case nodeconfig.Devnet:
		var dnConfig intelchainconfig.DevnetConfig
		if config.Devnet != nil {
			dnConfig = *config.Devnet
		} else {
			dnConfig = intelchainConfigs.GetDefaultDevnetConfigCopy()
		}

		devnetConfig, err := shardingconfig.NewInstance(
			uint32(dnConfig.NumShards), dnConfig.ShardSize,
			dnConfig.ItcNodeSize, dnConfig.SlotsLimit,
			numeric.OneDec(), genesis.IntelchainAccounts,
			genesis.FoundationalNodeAccounts, shardingconfig.Allowlist{},
			nil, numeric.ZeroDec(), ethCommon.Address{},
			nil, shardingconfig.VLBPE,
		)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ERROR invalid devnet sharding config: %s",
				err)
			os.Exit(1)
		}
		shard.Schedule = shardingconfig.NewFixedSchedule(devnetConfig)
	}
}

func findAccountsByPubKeys(config shardingconfig.Instance, pubKeys multibls.PublicKeys) {
	for _, key := range pubKeys {
		keyStr := key.Bytes.Hex()
		_, account := config.FindAccount(keyStr)
		if account != nil {
			initialAccounts = append(initialAccounts, account)
		}
	}
}

func setupLegacyNodeAccount(hc intelchainconfig.IntelchainConfig) error {
	genesisShardingConfig := shard.Schedule.InstanceForEpoch(big.NewInt(core.GenesisEpoch))
	multiBLSPubKey := setupConsensusKeys(hc, nodeconfig.GetDefaultConfig())

	reshardingEpoch := genesisShardingConfig.ReshardingEpoch()
	if len(reshardingEpoch) > 0 {
		for _, epoch := range reshardingEpoch {
			config := shard.Schedule.InstanceForEpoch(epoch)
			findAccountsByPubKeys(config, multiBLSPubKey)
			if len(initialAccounts) != 0 {
				break
			}
		}
	} else {
		findAccountsByPubKeys(genesisShardingConfig, multiBLSPubKey)
	}

	if len(initialAccounts) == 0 {
		fmt.Fprintf(
			os.Stderr,
			"ERROR cannot find your BLS key in the genesis/FN tables: %s\n",
			multiBLSPubKey.SerializeToHexStr(),
		)
		os.Exit(100)
	}

	for _, account := range initialAccounts {
		fmt.Printf("My Genesis Account: %v\n", *account)
	}
	return nil
}

func setupStakingNodeAccount(hc intelchainconfig.IntelchainConfig) error {
	pubKeys := setupConsensusKeys(hc, nodeconfig.GetDefaultConfig())
	shardID, err := nodeconfig.GetDefaultConfig().ShardIDFromConsensusKey()
	if err != nil {
		return errors.Wrap(err, "cannot determine shard to join")
	}
	if err := nodeconfig.GetDefaultConfig().ValidateConsensusKeysForSameShard(
		pubKeys, shardID,
	); err != nil {
		return err
	}
	for _, blsKey := range pubKeys {
		initialAccount := &genesis.DeployAccount{}
		initialAccount.ShardID = shardID
		initialAccount.BLSPublicKey = blsKey.Bytes.Hex()
		initialAccount.Address = ""
		initialAccounts = append(initialAccounts, initialAccount)
	}
	return nil
}

func createGlobalConfig(hc intelchainconfig.IntelchainConfig) (*nodeconfig.ConfigType, error) {
	var err error

	if len(initialAccounts) == 0 {
		initialAccounts = append(initialAccounts, &genesis.DeployAccount{ShardID: uint32(hc.General.ShardID)})
	}
	nodeConfig := nodeconfig.GetShardConfig(initialAccounts[0].ShardID)
	if hc.General.NodeType == intelchainConfigs.NodeTypeValidator {
		// Set up consensus keys.
		setupConsensusKeys(hc, nodeConfig)
	} else {
		// set dummy bls key for consensus object
		nodeConfig.ConsensusPriKey = multibls.GetPrivateKeys(&bls.SecretKey{})
	}

	// Set network type
	netType := nodeconfig.NetworkType(hc.Network.NetworkType)
	nodeconfig.SetNetworkType(netType)                // sets for both global and shard configs
	nodeConfig.SetShardID(initialAccounts[0].ShardID) // sets shard ID
	nodeConfig.SetArchival(hc.General.IsBeaconArchival, hc.General.IsArchival)
	nodeConfig.IsOffline = hc.General.IsOffline
	nodeConfig.Downloader = hc.Sync.Downloader
	nodeConfig.StagedSync = hc.Sync.StagedSync
	nodeConfig.StagedSyncTurboMode = hc.Sync.StagedSyncCfg.TurboMode
	nodeConfig.UseMemDB = hc.Sync.StagedSyncCfg.UseMemDB
	nodeConfig.DoubleCheckBlockHashes = hc.Sync.StagedSyncCfg.DoubleCheckBlockHashes
	nodeConfig.MaxBlocksPerSyncCycle = hc.Sync.StagedSyncCfg.MaxBlocksPerSyncCycle
	nodeConfig.MaxBackgroundBlocks = hc.Sync.StagedSyncCfg.MaxBackgroundBlocks
	nodeConfig.MaxMemSyncCycleSize = hc.Sync.StagedSyncCfg.MaxMemSyncCycleSize
	nodeConfig.VerifyAllSig = hc.Sync.StagedSyncCfg.VerifyAllSig
	nodeConfig.VerifyHeaderBatchSize = hc.Sync.StagedSyncCfg.VerifyHeaderBatchSize
	nodeConfig.InsertChainBatchSize = hc.Sync.StagedSyncCfg.InsertChainBatchSize
	nodeConfig.LogProgress = hc.Sync.StagedSyncCfg.LogProgress
	nodeConfig.DebugMode = hc.Sync.StagedSyncCfg.DebugMode
	// P2P private key is used for secure message transfer between p2p nodes.
	nodeConfig.P2PPriKey, _, err = utils.LoadKeyFromFile(hc.P2P.KeyFile)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load or create P2P key at %#v",
			hc.P2P.KeyFile)
	}

	selfPeer := p2p.Peer{
		IP:              hc.P2P.IP,
		Port:            strconv.Itoa(hc.P2P.Port),
		ConsensusPubKey: nodeConfig.ConsensusPriKey[0].Pub.Object,
	}

	// for local-net the node has to be forced to assume it is public reachable
	forceReachabilityPublic := false
	if hc.Network.NetworkType == nodeconfig.Localnet {
		forceReachabilityPublic = true
	}

	myHost, err = p2p.NewHost(p2p.HostConfig{
		Self:                     &selfPeer,
		BLSKey:                   nodeConfig.P2PPriKey,
		BootNodes:                hc.Network.BootNodes,
		DataStoreFile:            hc.P2P.DHTDataStore,
		DiscConcurrency:          hc.P2P.DiscConcurrency,
		MaxConnPerIP:             hc.P2P.MaxConnsPerIP,
		DisablePrivateIPScan:     hc.P2P.DisablePrivateIPScan,
		MaxPeers:                 hc.P2P.MaxPeers,
		ConnManagerLowWatermark:  hc.P2P.ConnManagerLowWatermark,
		ConnManagerHighWatermark: hc.P2P.ConnManagerHighWatermark,
		WaitForEachPeerToConnect: hc.P2P.WaitForEachPeerToConnect,
		ForceReachabilityPublic:  forceReachabilityPublic,
		NoTransportSecurity:      hc.P2P.NoTransportSecurity,
		NAT:                      hc.P2P.NAT,
		UserAgent:                hc.P2P.UserAgent,
		DialTimeout:              hc.P2P.DialTimeout,
		Muxer:                    hc.P2P.Muxer,
		NoRelay:                  hc.P2P.NoRelay,
	})
	if err != nil {
		return nil, errors.Wrap(err, "cannot create P2P network host")
	}

	nodeConfig.DBDir = hc.General.DataDir

	if hc.Legacy != nil && hc.Legacy.WebHookConfig != nil && len(*hc.Legacy.WebHookConfig) != 0 {
		p := *hc.Legacy.WebHookConfig
		config, err := webhooks.NewWebHooksFromPath(p)
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "yaml path is bad: %s", p,
			)
			os.Exit(1)
		}
		nodeConfig.WebHooks.Hooks = config
	}

	nodeConfig.NtpServer = hc.Sys.NtpServer

	nodeConfig.TraceEnable = hc.General.TraceEnable

	return nodeConfig, nil
}

func setupChain(hc intelchainconfig.IntelchainConfig, nodeConfig *nodeconfig.ConfigType, registry *registry.Registry) *registry.Registry {

	// Current node.
	var chainDBFactory shardchain.DBFactory
	if hc.General.RunElasticMode {
		chainDBFactory = setupTiKV(hc)
	} else if hc.ShardData.EnableShardData {
		chainDBFactory = &shardchain.LDBShardFactory{
			RootDir:    nodeConfig.DBDir,
			DiskCount:  hc.ShardData.DiskCount,
			ShardCount: hc.ShardData.ShardCount,
			CacheTime:  hc.ShardData.CacheTime,
			CacheSize:  hc.ShardData.CacheSize,
		}
	} else {
		chainDBFactory = &shardchain.LDBFactory{RootDir: nodeConfig.DBDir}
	}

	engine := chain.NewEngine()
	registry.SetEngine(engine)

	chainConfig := nodeConfig.GetNetworkType().ChainConfig()
	collection := shardchain.NewCollection(
		&hc, chainDBFactory, &core.GenesisInitializer{NetworkType: nodeConfig.GetNetworkType()}, engine, &chainConfig,
	)
	for shardID, archival := range nodeConfig.ArchiveModes() {
		if archival {
			collection.DisableCache(shardID)
		}
	}
	registry.SetShardChainCollection(collection)

	var blockchain core.BlockChain

	// We are not beacon chain, make sure beacon already initialized.
	if nodeConfig.ShardID != shard.BeaconChainShardID {
		beacon, err := collection.ShardChain(shard.BeaconChainShardID, core.Options{EpochChain: true})
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error :%v \n", err)
			os.Exit(1)
		}
		registry.SetBeaconchain(beacon)
	}

	blockchain, err := collection.ShardChain(nodeConfig.ShardID)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error :%v \n", err)
		os.Exit(1)
	}
	registry.SetBlockchain(blockchain)
	if registry.GetBeaconchain() == nil {
		registry.SetBeaconchain(registry.GetBlockchain())
	}
	return registry
}

func setupConsensusAndNode(hc intelchainconfig.IntelchainConfig, nodeConfig *nodeconfig.ConfigType, registry *registry.Registry) *node.Node {
	decider := quorum.NewDecider(quorum.SuperMajorityVote, uint32(hc.General.ShardID))

	// Parse minPeers from intelchainconfig.IntelchainConfig
	var minPeers int
	var aggregateSig bool
	if hc.Consensus != nil {
		minPeers = hc.Consensus.MinPeers
		aggregateSig = hc.Consensus.AggregateSig
	} else {
		defaultConsensusConfig := intelchainConfigs.GetDefaultConsensusConfigCopy()
		minPeers = defaultConsensusConfig.MinPeers
		aggregateSig = defaultConsensusConfig.AggregateSig
	}

	blacklist, err := setupBlacklist(hc)
	if err != nil {
		utils.Logger().Warn().Msgf("Blacklist setup error: %s", err.Error())
	}
	allowedTxs, err := setupAllowedTxs(hc)
	if err != nil {
		utils.Logger().Warn().Msgf("AllowedTxs setup error: %s", err.Error())
	}
	localAccounts, err := setupLocalAccounts(hc, blacklist)
	if err != nil {
		utils.Logger().Warn().Msgf("local accounts setup error: %s", err.Error())
	}

	registry = setupChain(hc, nodeConfig, registry)
	if registry.GetShardChainCollection() == nil {
		panic("shard chain collection is nil1111111")
	}
	registry.SetWebHooks(nodeConfig.WebHooks.Hooks)
	cxPool := core.NewCxPool(core.CxPoolSize)
	registry.SetCxPool(cxPool)

	// Consensus object.
	registry.SetIsBackup(isBackup(hc))
	currentConsensus, err := consensus.New(
		myHost, nodeConfig.ShardID, nodeConfig.ConsensusPriKey, registry, decider, minPeers, aggregateSig)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error :%v \n", err)
		os.Exit(1)
	}

	currentNode := node.New(myHost, currentConsensus, blacklist, allowedTxs, localAccounts, &hc, registry)

	if hc.Legacy != nil && hc.Legacy.TPBroadcastInvalidTxn != nil {
		currentNode.BroadcastInvalidTx = *hc.Legacy.TPBroadcastInvalidTxn
	} else {
		currentNode.BroadcastInvalidTx = intelchainConfigs.DefaultBroadcastInvalidTx
	}

	// Syncing provider is provided by following rules:
	//   1. If starting with a localnet or offline, use local sync peers.
	//   2. If specified with --dns=false, use legacy syncing which is syncing through self-
	//      discover peers.
	//   3. Else, use the dns for syncing.
	if hc.Network.NetworkType == nodeconfig.Localnet || hc.General.IsOffline {
		epochConfig := shard.Schedule.InstanceForEpoch(ethCommon.Big0)
		selfPort := hc.P2P.Port
		currentNode.SyncingPeerProvider = node.NewLocalSyncingPeerProvider(
			6000, uint16(selfPort), epochConfig.NumShards(), uint32(epochConfig.NumNodesPerShard()))
	} else {
		addrs := myHost.GetP2PHost().Addrs()
		currentNode.SyncingPeerProvider = node.NewDNSSyncingPeerProvider(hc.DNSSync.Zone, strconv.Itoa(hc.DNSSync.Port), addrs)
	}
	currentNode.NodeConfig.DNSZone = hc.DNSSync.Zone

	currentNode.NodeConfig.SetBeaconGroupID(
		nodeconfig.NewGroupIDByShardID(shard.BeaconChainShardID),
	)

	nodeconfig.GetDefaultConfig().DBDir = nodeConfig.DBDir
	processNodeType(hc, currentNode.NodeConfig)
	currentNode.NodeConfig.SetShardGroupID(nodeconfig.NewGroupIDByShardID(nodeconfig.ShardID(nodeConfig.ShardID)))
	currentNode.NodeConfig.SetClientGroupID(nodeconfig.NewClientGroupIDByShardID(shard.BeaconChainShardID))
	currentNode.NodeConfig.ConsensusPriKey = nodeConfig.ConsensusPriKey

	// This needs to be executed after consensus setup
	if err := currentConsensus.InitConsensusWithValidators(); err != nil {
		utils.Logger().Warn().
			Int("shardID", hc.General.ShardID).
			Err(err).
			Msg("InitConsensusWithMembers failed")
	}

	// Set the consensus ID to be the current block number
	viewID := currentNode.Blockchain().CurrentBlock().Header().ViewID().Uint64()
	currentConsensus.SetViewIDs(viewID + 1)
	utils.Logger().Info().
		Uint64("viewID", viewID).
		Msg("Init Blockchain")

	currentNode.Consensus.Registry().SetNodeConfig(currentNode.NodeConfig)
	// update consensus information based on the blockchain
	currentConsensus.SetMode(currentConsensus.UpdateConsensusInformation())
	currentConsensus.NextBlockDue = time.Now()
	return currentNode
}

func setupTiKV(hc intelchainconfig.IntelchainConfig) shardchain.DBFactory {
	err := redis_helper.Init(hc.TiKV.StateDBRedisServerAddr)
	if err != nil {
		panic("can not connect to redis: " + err.Error())
	}

	factory := &shardchain.TiKvFactory{
		PDAddr: hc.TiKV.PDAddr,
		Role:   hc.TiKV.Role,
		CacheConfig: statedb_cache.StateDBCacheConfig{
			CacheSizeInMB:        hc.TiKV.StateDBCacheSizeInMB,
			CachePersistencePath: hc.TiKV.StateDBCachePersistencePath,
			RedisServerAddr:      hc.TiKV.StateDBRedisServerAddr,
			RedisLRUTimeInDay:    hc.TiKV.StateDBRedisLRUTimeInDay,
			DebugHitRate:         hc.TiKV.Debug,
		},
	}

	tikv_manage.SetDefaultTiKVFactory(factory)
	return factory
}

func processNodeType(hc intelchainconfig.IntelchainConfig, nodeConfig *nodeconfig.ConfigType) {
	switch hc.General.NodeType {
	case intelchainConfigs.NodeTypeExplorer:
		nodeconfig.SetDefaultRole(nodeconfig.ExplorerNode)
		nodeConfig.SetRole(nodeconfig.ExplorerNode)

	case intelchainConfigs.NodeTypeValidator:
		nodeconfig.SetDefaultRole(nodeconfig.Validator)
		nodeConfig.SetRole(nodeconfig.Validator)
	}
}

func isBackup(hc intelchainconfig.IntelchainConfig) (isBackup bool) {
	switch hc.General.NodeType {
	case intelchainConfigs.NodeTypeExplorer:

	case intelchainConfigs.NodeTypeValidator:
		return hc.General.IsBackup
	}
	return false
}

func setupPprofService(node *node.Node, hc intelchainconfig.IntelchainConfig) {
	pprofConfig := pprof.Config{
		Enabled:            hc.Pprof.Enabled,
		ListenAddr:         hc.Pprof.ListenAddr,
		Folder:             hc.Pprof.Folder,
		ProfileNames:       hc.Pprof.ProfileNames,
		ProfileIntervals:   hc.Pprof.ProfileIntervals,
		ProfileDebugValues: hc.Pprof.ProfileDebugValues,
	}
	s := pprof.NewService(pprofConfig)
	node.RegisterService(service.Pprof, s)
}

func setupPrometheusService(node *node.Node, hc intelchainconfig.IntelchainConfig, sid uint32) {
	prometheusConfig := prometheus.Config{
		Enabled:    hc.Prometheus.Enabled,
		IP:         hc.Prometheus.IP,
		Port:       hc.Prometheus.Port,
		EnablePush: hc.Prometheus.EnablePush,
		Gateway:    hc.Prometheus.Gateway,
		Network:    hc.Network.NetworkType,
		Legacy:     hc.General.NoStaking,
		NodeType:   hc.General.NodeType,
		Shard:      sid,
		Instance:   myHost.GetID().String(),
	}

	if hc.General.RunElasticMode {
		prometheusConfig.TikvRole = hc.TiKV.Role
	}

	p := prometheus.NewService(prometheusConfig)
	node.RegisterService(service.Prometheus, p)
}

func setupSyncService(node *node.Node, host p2p.Host, hc intelchainconfig.IntelchainConfig) {
	blockchains := []core.BlockChain{node.Blockchain()}
	if node.Blockchain().ShardID() != shard.BeaconChainShardID {
		blockchains = append(blockchains, node.EpochChain())
	}

	dConfig := downloader.Config{
		ServerOnly:   !hc.Sync.Downloader,
		Network:      nodeconfig.NetworkType(hc.Network.NetworkType),
		Concurrency:  hc.Sync.Concurrency,
		MinStreams:   hc.Sync.MinPeers,
		InitStreams:  hc.Sync.InitStreams,
		SmSoftLowCap: hc.Sync.DiscSoftLowCap,
		SmHardLowCap: hc.Sync.DiscHardLowCap,
		SmHiCap:      hc.Sync.DiscHighCap,
		SmDiscBatch:  hc.Sync.DiscBatch,
	}
	// If we are running side chain, we will need to do some extra works for beacon
	// sync.
	if !node.IsRunningBeaconChain() {
		dConfig.BHConfig = &downloader.BeaconHelperConfig{
			BlockC:     node.BeaconBlockChannel,
			InsertHook: node.BeaconSyncHook,
		}
	}
	s := synchronize.NewService(host, blockchains, dConfig)

	node.RegisterService(service.Synchronize, s)

	d := s.Downloaders.GetShardDownloader(node.Blockchain().ShardID())
	if hc.Sync.Downloader && hc.General.NodeType != intelchainConfigs.NodeTypeExplorer {
		node.Consensus.SetDownloader(d) // Set downloader when stream client is active
	}
}

func setupStagedSyncService(node *node.Node, host p2p.Host, hc intelchainconfig.IntelchainConfig) {
	blockchains := []core.BlockChain{node.Blockchain()}
	if node.Blockchain().ShardID() != shard.BeaconChainShardID {
		blockchains = append(blockchains, node.EpochChain())
	}

	sConfig := stagedstreamsync.Config{
		ServerOnly:           !hc.Sync.Downloader,
		SyncMode:             stagedstreamsync.SyncMode(hc.Sync.SyncMode),
		Network:              nodeconfig.NetworkType(hc.Network.NetworkType),
		Concurrency:          hc.Sync.Concurrency,
		MinStreams:           hc.Sync.MinPeers,
		InitStreams:          hc.Sync.InitStreams,
		MaxAdvertiseWaitTime: hc.Sync.MaxAdvertiseWaitTime,
		SmSoftLowCap:         hc.Sync.DiscSoftLowCap,
		SmHardLowCap:         hc.Sync.DiscHardLowCap,
		SmHiCap:              hc.Sync.DiscHighCap,
		SmDiscBatch:          hc.Sync.DiscBatch,
		UseMemDB:             hc.Sync.StagedSyncCfg.UseMemDB,
		LogProgress:          hc.Sync.StagedSyncCfg.LogProgress,
		DebugMode:            true, // hc.Sync.StagedSyncCfg.DebugMode,
	}

	// If we are running side chain, we will need to do some extra works for beacon
	// sync.
	if !node.IsRunningBeaconChain() {
		sConfig.BHConfig = &stagedstreamsync.BeaconHelperConfig{
			BlockC:     node.BeaconBlockChannel,
			InsertHook: node.BeaconSyncHook,
		}
	}
	//Setup stream sync service
	s := stagedstreamsync.NewService(host, blockchains, node.Consensus, sConfig, hc.General.DataDir)

	node.RegisterService(service.StagedStreamSync, s)

	d := s.Downloaders.GetShardDownloader(node.Blockchain().ShardID())
	if hc.Sync.Downloader && hc.General.NodeType != intelchainConfigs.NodeTypeExplorer {
		node.Consensus.SetDownloader(d) // Set downloader when stream client is active
	}
}

func setupBlacklist(hc intelchainconfig.IntelchainConfig) (map[ethCommon.Address]struct{}, error) {
	rosetta_common.InitRosettaFile(hc.TxPool.RosettaFixFile)

	utils.Logger().Debug().Msgf("Using blacklist file at `%s`", hc.TxPool.BlacklistFile)
	dat, err := os.ReadFile(hc.TxPool.BlacklistFile)
	if err != nil {
		return nil, err
	}
	addrMap := make(map[ethCommon.Address]struct{})
	for _, line := range strings.Split(string(dat), "\n") {
		if len(line) != 0 { // blacklist file may have trailing empty string line
			b32 := strings.TrimSpace(strings.Split(string(line), "#")[0])
			addr, err := common.ParseAddr(b32)
			if err != nil {
				return nil, err
			}
			addrMap[addr] = struct{}{}
		}
	}
	return addrMap, nil
}

func parseAllowedTxs(data []byte) (map[ethCommon.Address][]core.AllowedTxData, error) {
	allowedTxs := make(map[ethCommon.Address][]core.AllowedTxData)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if len(line) != 0 { // AllowedTxs file may have trailing empty string line
			substrings := strings.Split(string(line), "->")
			fromStr := strings.TrimSpace(substrings[0])
			txSubstrings := strings.Split(substrings[1], ":")
			toStr := strings.TrimSpace(txSubstrings[0])
			dataStr := strings.TrimSpace(txSubstrings[1])
			from, err := common.ParseAddr(fromStr)
			if err != nil {
				return nil, err
			}
			to, err := common.ParseAddr(toStr)
			if err != nil {
				return nil, err
			}
			data, err := hexutil.Decode(dataStr)
			if err != nil {
				return nil, err
			}
			allowedTxs[from] = append(allowedTxs[from], core.AllowedTxData{
				To:   to,
				Data: data,
			})
		}
	}
	return allowedTxs, nil
}

func setupAllowedTxs(hc intelchainconfig.IntelchainConfig) (map[ethCommon.Address][]core.AllowedTxData, error) {
	// check if the file exists
	if _, err := os.Stat(hc.TxPool.AllowedTxsFile); err == nil {
		// read the file and parse allowed transactions
		utils.Logger().Debug().Msgf("Using AllowedTxs file at `%s`", hc.TxPool.AllowedTxsFile)
		data, err := os.ReadFile(hc.TxPool.AllowedTxsFile)
		if err != nil {
			return nil, err
		}
		return parseAllowedTxs(data)
	} else if errors.Is(err, os.ErrNotExist) {
		// file path does not exist
		utils.Logger().Debug().
			Str("AllowedTxsFile", hc.TxPool.AllowedTxsFile).
			Msg("AllowedTxs file doesn't exist")
		return make(map[ethCommon.Address][]core.AllowedTxData), nil
	} else {
		// some other errors happened
		utils.Logger().Error().Err(err).Msg("setup allowedTxs failed")
		return nil, err
	}
}

func setupLocalAccounts(hc intelchainconfig.IntelchainConfig, blacklist map[ethCommon.Address]struct{}) ([]ethCommon.Address, error) {
	file := hc.TxPool.LocalAccountsFile
	// check if file exist
	var fileData string
	if _, err := os.Stat(file); err == nil {
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		fileData = string(b)
	} else if errors.Is(err, os.ErrNotExist) {
		// file path does not exist
		return []ethCommon.Address{}, nil
	} else {
		// some other errors happened
		return nil, err
	}

	localAccounts := make(map[ethCommon.Address]struct{})
	lines := strings.Split(fileData, "\n")
	for _, line := range lines {
		if len(line) == 0 { // the file may have trailing empty string line
			continue
		}
		addrPart := strings.TrimSpace(strings.Split(string(line), "#")[0])
		if len(addrPart) == 0 { // if the line is commented by #
			continue
		}
		addr, err := common.ParseAddr(addrPart)
		if err != nil {
			return nil, err
		}
		// skip the blacklisted addresses
		if _, exists := blacklist[addr]; exists {
			utils.Logger().Warn().Msgf("local account with address %s is blacklisted", addr.String())
			continue
		}
		localAccounts[addr] = struct{}{}
	}
	uniqueAddresses := make([]ethCommon.Address, 0, len(localAccounts))
	for addr := range localAccounts {
		uniqueAddresses = append(uniqueAddresses, addr)
	}

	return uniqueAddresses, nil
}

func listenOSSigAndShutDown(node *node.Node) {
	// Prepare for graceful shutdown from os signals
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
	sig := <-osSignal
	utils.Logger().Warn().Str("signal", sig.String()).Msg("Gracefully shutting down...")
	const msg = "Got %s signal. Gracefully shutting down...\n"
	fmt.Fprintf(os.Stderr, msg, sig)

	go node.ShutDown()

	for i := 10; i > 0; i-- {
		<-osSignal
		if i > 1 {
			fmt.Printf("Already shutting down, interrupt more to force quit: (times=%v)\n", i-1)
		}
	}
	fmt.Println("Forced QUIT.")
	os.Exit(-1)
}
