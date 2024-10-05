package config

import (
	"time"

	"github.com/zennittians/intelchain/core"
	intelchainconfig "github.com/zennittians/intelchain/internal/configs/intelchain"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/itc"
)

const tomlConfigVersion = "2.6.2"

const (
	defNetworkType = nodeconfig.Mainnet
)

var defaultConfig = intelchainconfig.IntelchainConfig{
	Version: tomlConfigVersion,
	General: intelchainconfig.GeneralConfig{
		NodeType:         "validator",
		NoStaking:        false,
		ShardID:          -1,
		IsArchival:       false,
		IsBeaconArchival: false,
		IsOffline:        false,
		DataDir:          "./",
		TraceEnable:      false,
	},
	Network: GetDefaultNetworkConfig(defNetworkType),
	P2P: intelchainconfig.P2pConfig{
		Port:                     nodeconfig.DefaultP2PPort,
		IP:                       nodeconfig.DefaultPublicListenIP,
		KeyFile:                  "./.itckey",
		DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
		MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
		DisablePrivateIPScan:     false,
		MaxPeers:                 nodeconfig.DefaultMaxPeers,
		ConnManagerLowWatermark:  nodeconfig.DefaultConnManagerLowWatermark,
		ConnManagerHighWatermark: nodeconfig.DefaultConnManagerHighWatermark,
		WaitForEachPeerToConnect: nodeconfig.DefaultWaitForEachPeerToConnect,
		NoTransportSecurity:      nodeconfig.DefaultNoTransportSecurity,
		NAT:                      nodeconfig.DefaultNAT,
		UserAgent:                nodeconfig.DefaultUserAgent,
		DialTimeout:              nodeconfig.DefaultDialTimeout,
		Muxer:                    nodeconfig.DefaultMuxer,
		NoRelay:                  nodeconfig.DefaultNoRelay,
	},
	HTTP: intelchainconfig.HttpConfig{
		Enabled:        true,
		RosettaEnabled: false,
		IP:             "127.0.0.1",
		Port:           nodeconfig.DefaultRPCPort,
		AuthPort:       nodeconfig.DefaultAuthRPCPort,
		RosettaPort:    nodeconfig.DefaultRosettaPort,
		ReadTimeout:    nodeconfig.DefaultHTTPTimeoutRead,
		WriteTimeout:   nodeconfig.DefaultHTTPTimeoutWrite,
		IdleTimeout:    nodeconfig.DefaultHTTPTimeoutIdle,
	},
	WS: intelchainconfig.WsConfig{
		Enabled:  true,
		IP:       "127.0.0.1",
		Port:     nodeconfig.DefaultWSPort,
		AuthPort: nodeconfig.DefaultAuthWSPort,
	},
	RPCOpt: intelchainconfig.RpcOptConfig{
		DebugEnabled:       false,
		EthRPCsEnabled:     true,
		StakingRPCsEnabled: true,
		LegacyRPCsEnabled:  true,
		RpcFilterFile:      "./.itc/rpc_filter.txt",
		RateLimterEnabled:  true,
		RequestsPerSecond:  nodeconfig.DefaultRPCRateLimit,
		EvmCallTimeout:     nodeconfig.DefaultEvmCallTimeout,
		PreimagesEnabled:   false,
	},
	BLSKeys: intelchainconfig.BlsConfig{
		KeyDir:   "./.itc/blskeys",
		KeyFiles: []string{},
		MaxKeys:  10,

		PassEnabled:      true,
		PassSrcType:      blsPassTypeAuto,
		PassFile:         "",
		SavePassphrase:   false,
		KMSEnabled:       false,
		KMSConfigSrcType: kmsConfigTypeShared,
		KMSConfigFile:    "",
	},
	TxPool: intelchainconfig.TxPoolConfig{
		BlacklistFile:     "./.itc/blacklist.txt",
		AllowedTxsFile:    "./.itc/allowedtxs.txt",
		RosettaFixFile:    "",
		AccountSlots:      core.DefaultTxPoolConfig.AccountSlots,
		LocalAccountsFile: "./.itc/locals.txt",
		GlobalSlots:       core.DefaultTxPoolConfig.GlobalSlots,
		AccountQueue:      core.DefaultTxPoolConfig.AccountQueue,
		GlobalQueue:       core.DefaultTxPoolConfig.GlobalQueue,
		Lifetime:          core.DefaultTxPoolConfig.Lifetime,
		PriceLimit:        intelchainconfig.PriceLimit(core.DefaultTxPoolConfig.PriceLimit),
		PriceBump:         core.DefaultTxPoolConfig.PriceBump,
	},
	Sync: GetDefaultSyncConfig(defNetworkType),
	Pprof: intelchainconfig.PprofConfig{
		Enabled:            false,
		ListenAddr:         "127.0.0.1:6060",
		Folder:             "./profiles",
		ProfileNames:       []string{},
		ProfileIntervals:   []int{600},
		ProfileDebugValues: []int{0},
	},
	Log: intelchainconfig.LogConfig{
		Console:      false,
		Folder:       "./latest",
		FileName:     "intelchain.log",
		RotateSize:   100,
		RotateCount:  0,
		RotateMaxAge: 0,
		Verbosity:    3,
		VerbosePrints: intelchainconfig.LogVerbosePrints{
			Config: true,
		},
	},
	DNSSync: GetDefaultDNSSyncConfig(defNetworkType),
	ShardData: intelchainconfig.ShardDataConfig{
		EnableShardData: false,
		DiskCount:       8,
		ShardCount:      4,
		CacheTime:       10,
		CacheSize:       512,
	},
	GPO: intelchainconfig.GasPriceOracleConfig{
		Blocks:            itc.DefaultGPOConfig.Blocks,
		Transactions:      itc.DefaultGPOConfig.Transactions,
		Percentile:        itc.DefaultGPOConfig.Percentile,
		DefaultPrice:      itc.DefaultGPOConfig.DefaultPrice,
		MaxPrice:          itc.DefaultGPOConfig.MaxPrice,
		LowUsageThreshold: itc.DefaultGPOConfig.LowUsageThreshold,
		BlockGasLimit:     itc.DefaultGPOConfig.BlockGasLimit,
	},
	Cache: GetDefaultCacheConfig(defNetworkType),
}

var defaultSysConfig = intelchainconfig.SysConfig{
	NtpServer: "1.pool.ntp.org,0.beevik-ntp.pool.ntp.org",
}

var defaultDevnetConfig = intelchainconfig.DevnetConfig{
	NumShards:   2,
	ShardSize:   10,
	ItcNodeSize: 10,
	SlotsLimit:  0, // 0 means no limit
}

var defaultRevertConfig = intelchainconfig.RevertConfig{
	RevertBeacon: false,
	RevertBefore: 0,
	RevertTo:     0,
}

var defaultPreimageConfig = intelchainconfig.PreimageConfig{
	ImportFrom:    "",
	ExportTo:      "",
	GenerateStart: 0,
	GenerateEnd:   0,
}

var defaultLogContext = intelchainconfig.LogContext{
	IP:   "127.0.0.1",
	Port: 9000,
}

var defaultConsensusConfig = intelchainconfig.ConsensusConfig{
	MinPeers:     6,
	AggregateSig: true,
}

var defaultPrometheusConfig = intelchainconfig.PrometheusConfig{
	Enabled:    true,
	IP:         "0.0.0.0",
	Port:       9900,
	EnablePush: false,
	Gateway:    "https://gateway.intelchain.org",
}

var defaultStagedSyncConfig = intelchainconfig.StagedSyncConfig{
	TurboMode:              true,
	DoubleCheckBlockHashes: false,
	MaxBlocksPerSyncCycle:  512,   // sync new blocks in each cycle, if set to zero means all blocks in one full cycle
	MaxBackgroundBlocks:    512,   // max blocks to be downloaded at background process in turbo mode
	InsertChainBatchSize:   128,   // number of blocks to build a batch and insert to chain in staged sync
	VerifyAllSig:           false, // whether it should verify signatures for all blocks
	VerifyHeaderBatchSize:  100,   // batch size to verify block header before insert to chain
	MaxMemSyncCycleSize:    1024,  // max number of blocks to use a single transaction for staged sync
	UseMemDB:               true,  // it uses memory by default. set it to false to use disk
	LogProgress:            false, // log the full sync progress in console
	DebugMode:              false, // log every single process and error to help to debug the syncing (DebugMode is not accessible to the end user and is only an aid for development)
}

var (
	defaultMainnetSyncConfig = intelchainconfig.SyncConfig{
		Enabled:              false,
		SyncMode:             0,
		Downloader:           false,
		StagedSync:           false,
		StagedSyncCfg:        defaultStagedSyncConfig,
		Concurrency:          6,
		MinPeers:             6,
		InitStreams:          8,
		MaxAdvertiseWaitTime: 60, //minutes
		DiscSoftLowCap:       8,
		DiscHardLowCap:       6,
		DiscHighCap:          128,
		DiscBatch:            8,
	}

	defaultTestNetSyncConfig = intelchainconfig.SyncConfig{
		Enabled:              true,
		SyncMode:             0,
		Downloader:           false,
		StagedSync:           false,
		StagedSyncCfg:        defaultStagedSyncConfig,
		Concurrency:          2,
		MinPeers:             2,
		InitStreams:          2,
		MaxAdvertiseWaitTime: 5, //minutes
		DiscSoftLowCap:       2,
		DiscHardLowCap:       2,
		DiscHighCap:          1024,
		DiscBatch:            3,
	}

	defaultLocalNetSyncConfig = intelchainconfig.SyncConfig{
		Enabled:              true,
		SyncMode:             0,
		Downloader:           true,
		StagedSync:           true,
		StagedSyncCfg:        defaultStagedSyncConfig,
		Concurrency:          4,
		MinPeers:             4,
		InitStreams:          4,
		MaxAdvertiseWaitTime: 5, //minutes
		DiscSoftLowCap:       4,
		DiscHardLowCap:       4,
		DiscHighCap:          1024,
		DiscBatch:            8,
	}

	defaultPartnerSyncConfig = intelchainconfig.SyncConfig{
		Enabled:              false,
		SyncMode:             0,
		Downloader:           false,
		StagedSync:           false,
		StagedSyncCfg:        defaultStagedSyncConfig,
		Concurrency:          2,
		MinPeers:             2,
		InitStreams:          2,
		MaxAdvertiseWaitTime: 2, //minutes
		DiscSoftLowCap:       2,
		DiscHardLowCap:       2,
		DiscHighCap:          1024,
		DiscBatch:            4,
	}

	defaultElseSyncConfig = intelchainconfig.SyncConfig{
		Enabled:              false,
		SyncMode:             0,
		Downloader:           false,
		StagedSync:           false,
		StagedSyncCfg:        defaultStagedSyncConfig,
		Concurrency:          4,
		MinPeers:             4,
		InitStreams:          4,
		MaxAdvertiseWaitTime: 2, //minutes
		DiscSoftLowCap:       4,
		DiscHardLowCap:       4,
		DiscHighCap:          1024,
		DiscBatch:            8,
	}
)

var defaultCacheConfig = intelchainconfig.CacheConfig{
	Disabled:        false,
	TrieNodeLimit:   256,
	TriesInMemory:   128,
	TrieTimeLimit:   2 * time.Minute,
	SnapshotLimit:   0,
	SnapshotWait:    true,
	Preimages:       true,
	SnapshotNoBuild: false,
}

const (
	DefaultBroadcastInvalidTx = false
)

func GetDefaultConfigCopy() intelchainconfig.IntelchainConfig {
	config := defaultConfig
	return config
}

func GetDefaultItcConfigCopy(nt nodeconfig.NetworkType) intelchainconfig.IntelchainConfig {
	config := defaultConfig

	config.Network = GetDefaultNetworkConfig(nt)
	if nt == nodeconfig.Devnet {
		devnet := GetDefaultDevnetConfigCopy()
		config.Devnet = &devnet
	}
	config.Sync = GetDefaultSyncConfig(nt)
	config.DNSSync = GetDefaultDNSSyncConfig(nt)
	config.Cache = GetDefaultCacheConfig(nt)

	return config
}

func GetDefaultSysConfigCopy() intelchainconfig.SysConfig {
	config := defaultSysConfig
	return config
}

func GetDefaultDevnetConfigCopy() intelchainconfig.DevnetConfig {
	config := defaultDevnetConfig
	return config
}

func GetDefaultRevertConfigCopy() intelchainconfig.RevertConfig {
	config := defaultRevertConfig
	return config
}

func GetDefaultPreimageConfigCopy() intelchainconfig.PreimageConfig {
	config := defaultPreimageConfig
	return config
}

func GetDefaultLogContextCopy() intelchainconfig.LogContext {
	config := defaultLogContext
	return config
}

func GetDefaultConsensusConfigCopy() intelchainconfig.ConsensusConfig {
	config := defaultConsensusConfig
	return config
}

func GetDefaultPrometheusConfigCopy() intelchainconfig.PrometheusConfig {
	config := defaultPrometheusConfig
	return config
}

func GetDefaultCacheConfigCopy() intelchainconfig.CacheConfig {
	config := defaultCacheConfig
	return config
}

const (
	NodeTypeValidator = "validator"
	NodeTypeExplorer  = "explorer"
)

const (
	blsPassTypeAuto   = "auto"
	blsPassTypeFile   = "file"
	blsPassTypePrompt = "prompt"

	kmsConfigTypeShared = "shared"
	kmsConfigTypePrompt = "prompt"
	kmsConfigTypeFile   = "file"

	legacyBLSPassTypeDefault = "default"
	legacyBLSPassTypeStdin   = "stdin"
	legacyBLSPassTypeDynamic = "no-prompt"
	legacyBLSPassTypePrompt  = "prompt"
	legacyBLSPassTypeStatic  = "file"
	legacyBLSPassTypeNone    = "none"

	legacyBLSKmsTypeDefault = "default"
	legacyBLSKmsTypePrompt  = "prompt"
	legacyBLSKmsTypeFile    = "file"
	legacyBLSKmsTypeNone    = "none"
)
