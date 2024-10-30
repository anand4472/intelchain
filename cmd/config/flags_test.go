package config

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/zennittians/intelchain/common/denominations"
	intelchainconfig "github.com/zennittians/intelchain/internal/configs/intelchain"

	"github.com/spf13/cobra"

	"github.com/zennittians/intelchain/internal/cli"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
)

var (
	trueBool = true
)

func TestIntelchainFlags(t *testing.T) {
	tests := []struct {
		argStr    string
		expConfig intelchainconfig.IntelchainConfig
	}{
		{
			// running staking command from legacy node.sh
			argStr: "--bootnodes /ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv," +
				"/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9,/ip4/13.113.101." +
				"219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX,/ip4/99.81.170.167/tcp/12019/p" +
				"2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj --ip 8.8.8.8 --port 9000 --network_type=mainn" +
				"et --dns_zone=t.intelchain.org --blacklist=./.itc/blacklist.txt --min_peers=6 --max_bls_keys_per_node=" +
				"10 --broadcast_invalid_tx=true --verbosity=3 --is_archival=false --shard_id=-1 --staking=true -" +
				"-aws-config-source file:config.json --p2p.disc.concurrency 5 --p2p.security.max-conn-per-ip 5",
			expConfig: intelchainconfig.IntelchainConfig{
				Version: tomlConfigVersion,
				General: intelchainconfig.GeneralConfig{
					NodeType:   "validator",
					NoStaking:  false,
					ShardID:    -1,
					IsArchival: false,
					DataDir:    "./",
				},
				Network: intelchainconfig.NetworkConfig{
					NetworkType: "mainnet",
					BootNodes: []string{
						"/ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv",
						"/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9",
						"/ip4/13.113.101.219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX",
						"/ip4/99.81.170.167/tcp/12019/p2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj",
					},
				},
				DNSSync: intelchainconfig.DnsSync{
					Port:       6000,
					Zone:       "t.intelchain.org",
					Server:     true,
					Client:     true,
					ServerPort: nodeconfig.DefaultDNSPort,
				},
				P2P: intelchainconfig.P2pConfig{
					Port:                     9000,
					IP:                       defaultConfig.P2P.IP,
					KeyFile:                  defaultConfig.P2P.KeyFile,
					DiscConcurrency:          5,
					MaxConnsPerIP:            5,
					DisablePrivateIPScan:     false,
					MaxPeers:                 defaultConfig.P2P.MaxPeers,
					ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
					ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
					WaitForEachPeerToConnect: false,
					NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
					NAT:                      defaultConfig.P2P.NAT,
					UserAgent:                defaultConfig.P2P.UserAgent,
					DialTimeout:              defaultConfig.P2P.DialTimeout,
					Muxer:                    defaultConfig.P2P.Muxer,
					NoRelay:                  defaultConfig.P2P.NoRelay,
				},
				HTTP: intelchainconfig.HttpConfig{
					Enabled:        true,
					IP:             "127.0.0.1",
					Port:           9500,
					AuthPort:       9501,
					RosettaEnabled: false,
					RosettaPort:    9700,
					ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
					WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
					IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
				},
				RPCOpt: intelchainconfig.RpcOptConfig{
					DebugEnabled:       false,
					EthRPCsEnabled:     true,
					StakingRPCsEnabled: true,
					LegacyRPCsEnabled:  true,
					RpcFilterFile:      "./.itc/rpc_filter.txt",
					RateLimterEnabled:  true,
					RequestsPerSecond:  1000,
					EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
					PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
				},
				WS: intelchainconfig.WsConfig{
					Enabled:  true,
					IP:       "127.0.0.1",
					Port:     9800,
					AuthPort: 9801,
				},
				Consensus: &intelchainconfig.ConsensusConfig{
					MinPeers:     6,
					AggregateSig: true,
				},
				BLSKeys: intelchainconfig.BlsConfig{
					KeyDir:           "./.itc/blskeys",
					KeyFiles:         []string{},
					MaxKeys:          10,
					PassEnabled:      true,
					PassSrcType:      "auto",
					PassFile:         "",
					SavePassphrase:   false,
					KMSEnabled:       false,
					KMSConfigSrcType: "file",
					KMSConfigFile:    "config.json",
				},
				TxPool: intelchainconfig.TxPoolConfig{
					BlacklistFile:     "./.itc/blacklist.txt",
					AllowedTxsFile:    "./.itc/allowedtxs.txt",
					RosettaFixFile:    "",
					AccountSlots:      16,
					GlobalSlots:       4096,
					LocalAccountsFile: "./.itc/locals.txt",
					AccountQueue:      64,
					GlobalQueue:       5120,
					Lifetime:          30 * time.Minute,
					PriceLimit:        100e9,
					PriceBump:         1,
				},
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
					FileName:     "validator-8.8.8.8-9000.log",
					RotateSize:   100,
					RotateCount:  0,
					RotateMaxAge: 0,
					Verbosity:    3,
					Context: &intelchainconfig.LogContext{
						IP:   "8.8.8.8",
						Port: 9000,
					},
					VerbosePrints: intelchainconfig.LogVerbosePrints{
						Config: true,
					},
				},
				Sys: &intelchainconfig.SysConfig{
					NtpServer: defaultSysConfig.NtpServer,
				},
				Legacy: &intelchainconfig.LegacyConfig{
					TPBroadcastInvalidTxn: &trueBool,
				},
				Prometheus: &intelchainconfig.PrometheusConfig{
					Enabled:    true,
					IP:         "0.0.0.0",
					Port:       9900,
					EnablePush: true,
					Gateway:    "https://gateway.intelchain.org",
				},
				Sync: defaultMainnetSyncConfig,
				ShardData: intelchainconfig.ShardDataConfig{
					EnableShardData: false,
					DiskCount:       8,
					ShardCount:      4,
					CacheTime:       10,
					CacheSize:       512,
				},
				GPO: intelchainconfig.GasPriceOracleConfig{
					Blocks:            defaultConfig.GPO.Blocks,
					Transactions:      defaultConfig.GPO.Transactions,
					Percentile:        defaultConfig.GPO.Percentile,
					DefaultPrice:      defaultConfig.GPO.DefaultPrice,
					MaxPrice:          defaultConfig.GPO.MaxPrice,
					LowUsageThreshold: defaultConfig.GPO.LowUsageThreshold,
					BlockGasLimit:     defaultConfig.GPO.BlockGasLimit,
				},
				Cache: intelchainconfig.CacheConfig{
					Disabled:        defaultConfig.Cache.Disabled,
					TrieNodeLimit:   defaultCacheConfig.TrieNodeLimit,
					TriesInMemory:   defaultConfig.Cache.TriesInMemory,
					TrieTimeLimit:   defaultConfig.Cache.TrieTimeLimit,
					SnapshotLimit:   defaultConfig.Cache.SnapshotLimit,
					SnapshotWait:    defaultConfig.Cache.SnapshotWait,
					Preimages:       defaultConfig.Cache.Preimages,
					SnapshotNoBuild: defaultConfig.Cache.SnapshotNoBuild,
				},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, getRootFlags(), applyRootFlags)
		hc, err := ts.run(strings.Split(test.argStr, " "))
		if err != nil {
			t.Fatalf("Test %v: %v", i, err)
		}
		if !reflect.DeepEqual(hc, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, hc, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestGeneralFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.GeneralConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "validator",
				NoStaking:  false,
				ShardID:    -1,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.legacy", "--run.shard=0",
				"--run.archive=true", "--datadir=./.itc"},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  true,
				ShardID:    0,
				IsArchival: true,
				DataDir:    "./.itc",
			},
		},
		{
			args: []string{"--node_type", "explorer", "--staking", "--shard_id", "0",
				"--is_archival", "--db_dir", "./"},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: true,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--staking=false", "--is_archival=false"},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "validator",
				NoStaking:  true,
				ShardID:    -1,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.shard", "0"},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: false,
				DataDir:    "./",
			},
		},
		{
			args: []string{"--run", "explorer", "--run.shard", "0", "--run.archive=false"},
			expConfig: intelchainconfig.GeneralConfig{
				NodeType:   "explorer",
				NoStaking:  false,
				ShardID:    0,
				IsArchival: false,
				DataDir:    "./",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, generalFlags, applyGeneralFlags)

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.General, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, got.General, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestNetworkFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.IntelchainConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: intelchainconfig.IntelchainConfig{
				Network: intelchainconfig.NetworkConfig{
					NetworkType: defNetworkType,
					BootNodes:   nodeconfig.GetDefaultBootNodes(defNetworkType),
				},
				DNSSync: GetDefaultDNSSyncConfig(defNetworkType)},
		},
		{
			args: []string{"-n", "stn"},
			expConfig: intelchainconfig.IntelchainConfig{
				Network: intelchainconfig.NetworkConfig{
					NetworkType: nodeconfig.Stressnet,
					BootNodes:   nodeconfig.GetDefaultBootNodes(nodeconfig.Stressnet),
				},
				DNSSync: GetDefaultDNSSyncConfig(nodeconfig.Stressnet),
			},
		},
		{
			args: []string{"--network", "stk", "--bootnodes", "1,2,3,4", "--dns.zone", "8.8.8.8",
				"--dns.port", "9001", "--dns.server-port", "9002"},
			expConfig: intelchainconfig.IntelchainConfig{
				Network: intelchainconfig.NetworkConfig{
					NetworkType: "pangaea",
					BootNodes:   []string{"1", "2", "3", "4"},
				},
				DNSSync: intelchainconfig.DnsSync{
					Port:       9001,
					Zone:       "8.8.8.8",
					Server:     true,
					ServerPort: 9002,
				},
			},
		},
		{
			args: []string{"--network_type", "stk", "--bootnodes", "1,2,3,4", "--dns_zone", "8.8.8.8",
				"--dns_port", "9001"},
			expConfig: intelchainconfig.IntelchainConfig{
				Network: intelchainconfig.NetworkConfig{
					NetworkType: "pangaea",
					BootNodes:   []string{"1", "2", "3", "4"},
				},
				DNSSync: intelchainconfig.DnsSync{
					Port:       9001,
					Zone:       "8.8.8.8",
					Server:     true,
					ServerPort: nodeconfig.GetDefaultDNSPort(nodeconfig.Pangaea),
				},
			},
		},
	}
	for i, test := range tests {
		neededFlags := make([]cli.Flag, 0)
		neededFlags = append(neededFlags, networkFlags...)
		neededFlags = append(neededFlags, dnsSyncFlags...)
		ts := newFlagTestSuite(t, neededFlags, func(cmd *cobra.Command, config *intelchainconfig.IntelchainConfig) {
			// This is the network related logic in function getintelchainconfig.IntelchainConfig
			nt := getNetworkType(cmd)
			config.Network = GetDefaultNetworkConfig(nt)
			config.DNSSync = GetDefaultDNSSyncConfig(nt)
			applyNetworkFlags(cmd, config)
			applyDNSSyncFlags(cmd, config)
		})

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.Network, test.expConfig.Network) {
			t.Errorf("Test %v: unexpected network config: \n\t%+v\n\t%+v", i, got.Network, test.expConfig.Network)
		}
		if !reflect.DeepEqual(got.DNSSync, test.expConfig.DNSSync) {
			t.Errorf("Test %v: unexpected dnssync config: \n\t%+v\n\t%+v", i, got.DNSSync, test.expConfig.DNSSync)
		}
		ts.tearDown()
	}
}

var defDataStore = ".dht-127.0.0.1"

func TestP2PFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.P2pConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.P2P,
		},
		{
			args: []string{"--p2p.port", "9001", "--p2p.keyfile", "./key.file", "--p2p.dht.datastore",
				defDataStore},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     9001,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./key.file",
				DHTDataStore:             &defDataStore,
				MaxConnsPerIP:            10,
				DisablePrivateIPScan:     false,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--port", "9001", "--key", "./key.file"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     9001,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./key.file",
				MaxConnsPerIP:            10,
				DisablePrivateIPScan:     false,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.port", "9001", "--p2p.disc.concurrency", "5", "--p2p.security.max-conn-per-ip", "5"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     9001,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          5,
				MaxConnsPerIP:            5,
				DisablePrivateIPScan:     false,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.no-private-ip-scan"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     true,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.security.max-peers", "100"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 100,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.connmgr-low", "100"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  100,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.connmgr-high", "400"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: 400,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.connmgr-high", "400"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: 400,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.no-transport-security=false"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      false,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.nat=false"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      false,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.user-agent", "test1"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                "test1",
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.muxer", "mplex"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    "mplex",
				NoRelay:                  defaultConfig.P2P.NoRelay,
			},
		},
		{
			args: []string{"--p2p.no-relay=false"},
			expConfig: intelchainconfig.P2pConfig{
				Port:                     nodeconfig.DefaultP2PPort,
				IP:                       nodeconfig.DefaultPublicListenIP,
				KeyFile:                  "./.itckey",
				DiscConcurrency:          nodeconfig.DefaultP2PConcurrency,
				MaxConnsPerIP:            nodeconfig.DefaultMaxConnPerIP,
				DisablePrivateIPScan:     defaultConfig.P2P.DisablePrivateIPScan,
				MaxPeers:                 defaultConfig.P2P.MaxPeers,
				ConnManagerLowWatermark:  defaultConfig.P2P.ConnManagerLowWatermark,
				ConnManagerHighWatermark: defaultConfig.P2P.ConnManagerHighWatermark,
				WaitForEachPeerToConnect: false,
				NoTransportSecurity:      defaultConfig.P2P.NoTransportSecurity,
				NAT:                      defaultConfig.P2P.NAT,
				UserAgent:                defaultConfig.P2P.UserAgent,
				DialTimeout:              defaultConfig.P2P.DialTimeout,
				Muxer:                    defaultConfig.P2P.Muxer,
				NoRelay:                  false,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(p2pFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *intelchainconfig.IntelchainConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyP2PFlags(cmd, config)
			},
		)

		got, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(got.P2P, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, got.P2P, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRPCFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.HttpConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.HTTP,
		},
		{
			args: []string{"--http=false"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        false,
				RosettaEnabled: false,
				IP:             defaultConfig.HTTP.IP,
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.port", "9001"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             "8.8.8.8",
				Port:           9001,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.auth-port", "9001"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             "8.8.8.8",
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       9001,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.port", "9001", "--http.rosetta.port", "10001"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: true,
				IP:             "8.8.8.8",
				Port:           9001,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    10001,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--http.ip", "8.8.8.8", "--http.rosetta.port", "10001"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: true,
				IP:             "8.8.8.8",
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    10001,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--ip", "8.8.8.8", "--port", "9001", "--public_rpc"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             nodeconfig.DefaultPublicListenIP,
				Port:           9501,
				AuthPort:       9502,
				RosettaPort:    9701,
				ReadTimeout:    defaultConfig.HTTP.ReadTimeout,
				WriteTimeout:   defaultConfig.HTTP.WriteTimeout,
				IdleTimeout:    defaultConfig.HTTP.IdleTimeout,
			},
		},
		{
			args: []string{"--http.timeout.read", "10s", "--http.timeout.write", "20s", "--http.timeout.idle", "30s"},
			expConfig: intelchainconfig.HttpConfig{
				Enabled:        true,
				RosettaEnabled: false,
				IP:             defaultConfig.HTTP.IP,
				Port:           defaultConfig.HTTP.Port,
				AuthPort:       defaultConfig.HTTP.AuthPort,
				RosettaPort:    defaultConfig.HTTP.RosettaPort,
				ReadTimeout:    "10s",
				WriteTimeout:   "20s",
				IdleTimeout:    "30s",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(httpFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *intelchainconfig.IntelchainConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyHTTPFlags(cmd, config)
			},
		)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.HTTP, test.expConfig) {
			t.Errorf("Test %v: unexpected config: \n\t%+v\n\t%+v", i, hc.HTTP, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestWSFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.WsConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.WS,
		},
		{
			args: []string{"--ws=false"},
			expConfig: intelchainconfig.WsConfig{
				Enabled:  false,
				IP:       defaultConfig.WS.IP,
				Port:     defaultConfig.WS.Port,
				AuthPort: defaultConfig.WS.AuthPort,
			},
		},
		{
			args: []string{"--ws", "--ws.ip", "8.8.8.8", "--ws.port", "9001"},
			expConfig: intelchainconfig.WsConfig{
				Enabled:  true,
				IP:       "8.8.8.8",
				Port:     9001,
				AuthPort: defaultConfig.WS.AuthPort,
			},
		},
		{
			args: []string{"--ws", "--ws.ip", "8.8.8.8", "--ws.auth-port", "9001"},
			expConfig: intelchainconfig.WsConfig{
				Enabled:  true,
				IP:       "8.8.8.8",
				Port:     defaultConfig.WS.Port,
				AuthPort: 9001,
			},
		},
		{
			args: []string{"--ip", "8.8.8.8", "--port", "9001", "--public_rpc"},
			expConfig: intelchainconfig.WsConfig{
				Enabled:  true,
				IP:       nodeconfig.DefaultPublicListenIP,
				Port:     9801,
				AuthPort: 9802,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(wsFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *intelchainconfig.IntelchainConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyWSFlags(cmd, config)
			},
		)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.WS, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.WS, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRPCOptFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.RpcOptConfig
	}{
		{
			args: []string{"--rpc.debug"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       true,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.eth=false"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     false,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.staking=false"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: false,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.legacy=false"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  false,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.filterspath=./rmf.toml"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./rmf.toml",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.ratelimiter", "--rpc.ratelimit", "2000"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  2000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.ratelimiter=false", "--rpc.ratelimit", "2000"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  false,
				RequestsPerSecond:  2000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.evm-call-timeout", "10s"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     "10s",
				PreimagesEnabled:   defaultConfig.RPCOpt.PreimagesEnabled,
			},
		},

		{
			args: []string{"--rpc.preimages"},
			expConfig: intelchainconfig.RpcOptConfig{
				DebugEnabled:       false,
				EthRPCsEnabled:     true,
				StakingRPCsEnabled: true,
				LegacyRPCsEnabled:  true,
				RpcFilterFile:      "./.itc/rpc_filter.txt",
				RateLimterEnabled:  true,
				RequestsPerSecond:  1000,
				EvmCallTimeout:     defaultConfig.RPCOpt.EvmCallTimeout,
				PreimagesEnabled:   true,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, rpcOptFlags, applyRPCOptFlags)

		hc, _ := ts.run(test.args)

		if !reflect.DeepEqual(hc.RPCOpt, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.RPCOpt, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestBLSFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.BlsConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.BLSKeys,
		},
		{
			args: []string{"--bls.dir", "./blskeys", "--bls.keys", "key1,key2",
				"--bls.maxkeys", "8", "--bls.pass", "--bls.pass.src", "auto", "--bls.pass.save",
				"--bls.kms", "--bls.kms.src", "shared",
			},
			expConfig: intelchainconfig.BlsConfig{
				KeyDir:           "./blskeys",
				KeyFiles:         []string{"key1", "key2"},
				MaxKeys:          8,
				PassEnabled:      true,
				PassSrcType:      "auto",
				PassFile:         "",
				SavePassphrase:   true,
				KMSEnabled:       true,
				KMSConfigSrcType: "shared",
				KMSConfigFile:    "",
			},
		},
		{
			args: []string{"--bls.pass.file", "xxx.pass", "--bls.kms.config", "config.json"},
			expConfig: intelchainconfig.BlsConfig{
				KeyDir:           defaultConfig.BLSKeys.KeyDir,
				KeyFiles:         defaultConfig.BLSKeys.KeyFiles,
				MaxKeys:          defaultConfig.BLSKeys.MaxKeys,
				PassEnabled:      true,
				PassSrcType:      "file",
				PassFile:         "xxx.pass",
				SavePassphrase:   false,
				KMSEnabled:       false,
				KMSConfigSrcType: "file",
				KMSConfigFile:    "config.json",
			},
		},
		{
			args: []string{"--blskey_file", "key1,key2", "--blsfolder", "./itckeys",
				"--max_bls_keys_per_node", "5", "--blspass", "file:xxx.pass", "--save-passphrase",
				"--aws-config-source", "file:config.json",
			},
			expConfig: intelchainconfig.BlsConfig{
				KeyDir:           "./itckeys",
				KeyFiles:         []string{"key1", "key2"},
				MaxKeys:          5,
				PassEnabled:      true,
				PassSrcType:      "file",
				PassFile:         "xxx.pass",
				SavePassphrase:   true,
				KMSEnabled:       false,
				KMSConfigSrcType: "file",
				KMSConfigFile:    "config.json",
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, blsFlags, applyBLSFlags)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.BLSKeys, test.expConfig) {
			t.Errorf("Test %v: \n\t%+v\n\t%+v", i, hc.BLSKeys, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestConsensusFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *intelchainconfig.ConsensusConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--consensus.min-peers", "10", "--consensus.aggregate-sig=false"},
			expConfig: &intelchainconfig.ConsensusConfig{
				MinPeers:     10,
				AggregateSig: false,
			},
		},
		{
			args: []string{"--delay_commit", "10ms", "--block_period", "5", "--min_peers", "10",
				"--consensus.aggregate-sig=true"},
			expConfig: &intelchainconfig.ConsensusConfig{
				MinPeers:     10,
				AggregateSig: true,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, consensusFlags, applyConsensusFlags)

		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Consensus, test.expConfig) {
			t.Errorf("Test %v: unexpected config \n\t%+v\n\t%+v", i, hc.Consensus, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestTxPoolFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.TxPoolConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     defaultConfig.TxPool.BlacklistFile,
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				RosettaFixFile:    defaultConfig.TxPool.RosettaFixFile,
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--txpool.blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file", "--txpool.allowedtxs", "allowedtxs.txt"},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     "blacklist.file",
				AllowedTxsFile:    "allowedtxs.txt",
				RosettaFixFile:    "rosettafix.file",
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file"},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     "blacklist.file",
				RosettaFixFile:    "rosettafix.file",
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--txpool.accountslots", "5", "--txpool.blacklist", "blacklist.file", "--txpool.rosettafixfile", "rosettafix.file"},
			expConfig: intelchainconfig.TxPoolConfig{
				AccountSlots:      5,
				BlacklistFile:     "blacklist.file",
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				RosettaFixFile:    "rosettafix.file",
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--txpool.locals", "locals.txt"},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     defaultConfig.TxPool.BlacklistFile,
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				RosettaFixFile:    defaultConfig.TxPool.RosettaFixFile,
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				LocalAccountsFile: "locals.txt",
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--txpool.globalslots", "10240"},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     defaultConfig.TxPool.BlacklistFile,
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				RosettaFixFile:    defaultConfig.TxPool.RosettaFixFile,
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				GlobalSlots:       10240,
				AccountQueue:      defaultConfig.TxPool.AccountQueue,
				GlobalQueue:       defaultConfig.TxPool.GlobalQueue,
				Lifetime:          defaultConfig.TxPool.Lifetime,
				PriceLimit:        100e9,
				PriceBump:         1,
			},
		},
		{
			args: []string{"--txpool.accountqueue", "128", "--txpool.globalqueue", "10240", "--txpool.lifetime", "15m", "--txpool.pricelimit", "100", "--txpool.pricebump", "2"},
			expConfig: intelchainconfig.TxPoolConfig{
				BlacklistFile:     defaultConfig.TxPool.BlacklistFile,
				AllowedTxsFile:    defaultConfig.TxPool.AllowedTxsFile,
				RosettaFixFile:    defaultConfig.TxPool.RosettaFixFile,
				AccountSlots:      defaultConfig.TxPool.AccountSlots,
				LocalAccountsFile: defaultConfig.TxPool.LocalAccountsFile,
				GlobalSlots:       defaultConfig.TxPool.GlobalSlots,
				AccountQueue:      128,
				GlobalQueue:       10240,
				Lifetime:          15 * time.Minute,
				PriceLimit:        100,
				PriceBump:         2,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, txPoolFlags, applyTxPoolFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.TxPool, test.expConfig) {
			t.Errorf("Test %v: unexpected config\n\t%+v\n\t%+v", i, hc.TxPool, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestPprofFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.PprofConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.Pprof,
		},
		{
			args: []string{"--pprof"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.addr", "8.8.8.8:9001"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         "8.8.8.8:9001",
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof=false", "--pprof.addr", "8.8.8.8:9001"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            false,
				ListenAddr:         "8.8.8.8:9001",
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.names", "cpu,heap,mutex"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       []string{"cpu", "heap", "mutex"},
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.intervals", "0,1"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   []int{0, 1},
				ProfileDebugValues: defaultConfig.Pprof.ProfileDebugValues,
			},
		},
		{
			args: []string{"--pprof.profile.debug", "0,1,0"},
			expConfig: intelchainconfig.PprofConfig{
				Enabled:            true,
				ListenAddr:         defaultConfig.Pprof.ListenAddr,
				Folder:             defaultConfig.Pprof.Folder,
				ProfileNames:       defaultConfig.Pprof.ProfileNames,
				ProfileIntervals:   defaultConfig.Pprof.ProfileIntervals,
				ProfileDebugValues: []int{0, 1, 0},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, pprofFlags, applyPprofFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Pprof, test.expConfig) {
			t.Errorf("Test %v: unexpected config\n\t%+v\n\t%+v", i, hc.Pprof, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestLogFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.LogConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.Log,
		},
		{
			args: []string{"--log.dir", "latest_log", "--log.max-size", "10", "--log.rotate-count", "3",
				"--log.rotate-max-age", "0", "--log.name", "intelchain.log", "--log.verb", "5",
				"--log.verbose-prints", "config"},
			expConfig: intelchainconfig.LogConfig{
				Folder:       "latest_log",
				FileName:     "intelchain.log",
				RotateSize:   10,
				RotateCount:  3,
				RotateMaxAge: 0,
				Verbosity:    5,
				VerbosePrints: intelchainconfig.LogVerbosePrints{
					Config: true,
				},
				Context: nil,
			},
		},
		{
			args: []string{"--log.ctx.ip", "8.8.8.8", "--log.ctx.port", "9001"},
			expConfig: intelchainconfig.LogConfig{
				Console:       defaultConfig.Log.Console,
				Folder:        defaultConfig.Log.Folder,
				FileName:      defaultConfig.Log.FileName,
				RotateSize:    defaultConfig.Log.RotateSize,
				RotateCount:   defaultConfig.Log.RotateCount,
				RotateMaxAge:  defaultConfig.Log.RotateMaxAge,
				Verbosity:     defaultConfig.Log.Verbosity,
				VerbosePrints: defaultConfig.Log.VerbosePrints,
				Context: &intelchainconfig.LogContext{
					IP:   "8.8.8.8",
					Port: 9001,
				},
			},
		},
		{
			args: []string{"--log_folder", "latest_log", "--log_max_size", "10", "--verbosity",
				"5", "--ip", "8.8.8.8", "--port", "9001"},
			expConfig: intelchainconfig.LogConfig{
				Folder:        "latest_log",
				FileName:      "validator-8.8.8.8-9001.log",
				RotateSize:    10,
				RotateCount:   0,
				RotateMaxAge:  0,
				Verbosity:     5,
				VerbosePrints: defaultConfig.Log.VerbosePrints,
				Context: &intelchainconfig.LogContext{
					IP:   "8.8.8.8",
					Port: 9001,
				},
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, append(logFlags, legacyMiscFlags...),
			func(cmd *cobra.Command, config *intelchainconfig.IntelchainConfig) {
				applyLegacyMiscFlags(cmd, config)
				applyLogFlags(cmd, config)
			},
		)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Log, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Log, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestSysFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *intelchainconfig.SysConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: &intelchainconfig.SysConfig{
				NtpServer: defaultSysConfig.NtpServer,
			},
		},
		{
			args: []string{"--sys.ntp", "0.pool.ntp.org"},
			expConfig: &intelchainconfig.SysConfig{
				NtpServer: "0.pool.ntp.org",
			},
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, sysFlags, applySysFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Sys, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Sys, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestGPOFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.GasPriceOracleConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: intelchainconfig.GasPriceOracleConfig{
				Blocks:            defaultConfig.GPO.Blocks,
				Transactions:      defaultConfig.GPO.Transactions,
				Percentile:        defaultConfig.GPO.Percentile,
				DefaultPrice:      defaultConfig.GPO.DefaultPrice,
				MaxPrice:          defaultConfig.GPO.MaxPrice,
				LowUsageThreshold: defaultConfig.GPO.LowUsageThreshold,
				BlockGasLimit:     defaultConfig.GPO.BlockGasLimit,
			},
		},
		{
			args: []string{"--gpo.blocks", "5", "--gpo.transactions", "1", "--gpo.percentile", "2", "--gpo.defaultprice", "101000000000", "--gpo.maxprice", "400000000000", "--gpo.low-usage-threshold", "60", "--gpo.block-gas-limit", "10000000"},
			expConfig: intelchainconfig.GasPriceOracleConfig{
				Blocks:            5,
				Transactions:      1,
				Percentile:        2,
				DefaultPrice:      101 * denominations.Intello,
				MaxPrice:          400 * denominations.Intello,
				LowUsageThreshold: 60,
				BlockGasLimit:     10_000_000,
			},
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, gpoFlags, applyGPOFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.GPO, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.GPO, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestCacheFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.CacheConfig
		expErr    error
	}{
		{
			args: []string{},
			expConfig: intelchainconfig.CacheConfig{
				Disabled:        true, // based on network type
				TrieNodeLimit:   defaultCacheConfig.TrieNodeLimit,
				TriesInMemory:   defaultCacheConfig.TriesInMemory,
				TrieTimeLimit:   defaultCacheConfig.TrieTimeLimit,
				SnapshotLimit:   0, // based on network type,
				SnapshotWait:    defaultCacheConfig.SnapshotWait,
				Preimages:       defaultCacheConfig.Preimages, // based on network type
				SnapshotNoBuild: defaultCacheConfig.SnapshotNoBuild,
			},
		},
		{
			args: []string{"--cache.disabled=true", "--cache.trie_node_limit", "512", "--cache.tries_in_memory", "256", "--cache.preimages=false", "--cache.snapshot_limit", "512", "--cache.snapshot_no_build=true", "--cache.snapshot_wait=false"},
			expConfig: intelchainconfig.CacheConfig{
				Disabled:        true,
				TrieNodeLimit:   512,
				TriesInMemory:   256,
				TrieTimeLimit:   2 * time.Minute,
				SnapshotLimit:   512,
				SnapshotWait:    false,
				Preimages:       false,
				SnapshotNoBuild: true,
			},
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, cacheConfigFlags, applyCacheFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Cache, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Cache, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestDevnetFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *intelchainconfig.DevnetConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--devnet.num-shard", "3", "--devnet.shard-size", "100",
				"--devnet.itc-node-size", "60"},
			expConfig: &intelchainconfig.DevnetConfig{
				NumShards:   3,
				ShardSize:   100,
				ItcNodeSize: 60,
			},
		},
		{
			args: []string{"--dn_num_shards", "3", "--dn_shard_size", "100", "--dn_itc_size",
				"60"},
			expConfig: &intelchainconfig.DevnetConfig{
				NumShards:   3,
				ShardSize:   100,
				ItcNodeSize: 60,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, devnetFlags, applyDevnetFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}

		if !reflect.DeepEqual(hc.Devnet, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Devnet, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestRevertFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *intelchainconfig.RevertConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--revert.beacon"},
			expConfig: &intelchainconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     defaultRevertConfig.RevertTo,
				RevertBefore: defaultRevertConfig.RevertBefore,
			},
		},
		{
			args: []string{"--revert.beacon", "--revert.to", "100", "--revert.do-before", "10000"},
			expConfig: &intelchainconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     100,
				RevertBefore: 10000,
			},
		},
		{
			args: []string{"--revert_beacon", "--do_revert_before", "10000", "--revert_to", "100"},
			expConfig: &intelchainconfig.RevertConfig{
				RevertBeacon: true,
				RevertTo:     100,
				RevertBefore: 10000,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, revertFlags, applyRevertFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Revert, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Revert, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestPreimageFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig *intelchainconfig.PreimageConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: nil,
		},
		{
			args: []string{"--preimage.import", "/path/to/source.csv"},
			expConfig: &intelchainconfig.PreimageConfig{
				ImportFrom:    "/path/to/source.csv",
				ExportTo:      defaultPreimageConfig.ExportTo,
				GenerateStart: defaultPreimageConfig.GenerateStart,
				GenerateEnd:   defaultPreimageConfig.GenerateEnd,
			},
		},
		{
			args: []string{"--preimage.export", "/path/to/destination.csv"},
			expConfig: &intelchainconfig.PreimageConfig{
				ImportFrom:    defaultPreimageConfig.ImportFrom,
				ExportTo:      "/path/to/destination.csv",
				GenerateStart: defaultPreimageConfig.GenerateStart,
				GenerateEnd:   defaultPreimageConfig.GenerateEnd,
			},
		},
		{
			args: []string{"--preimage.start", "1"},
			expConfig: &intelchainconfig.PreimageConfig{
				ImportFrom:    defaultPreimageConfig.ImportFrom,
				ExportTo:      defaultPreimageConfig.ExportTo,
				GenerateStart: 1,
				GenerateEnd:   defaultPreimageConfig.GenerateEnd,
			},
		},
		{
			args: []string{"--preimage.end", "2"},
			expConfig: &intelchainconfig.PreimageConfig{
				ImportFrom:    defaultPreimageConfig.ImportFrom,
				ExportTo:      defaultPreimageConfig.ExportTo,
				GenerateStart: defaultPreimageConfig.GenerateStart,
				GenerateEnd:   2,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, preimageFlags, applyPreimageFlags)
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Preimage, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Preimage, test.expConfig)
		}
		ts.tearDown()
	}
}

func TestDNSSyncFlags(t *testing.T) {
	tests := []struct {
		args      []string
		network   string
		expConfig intelchainconfig.DnsSync
		expErr    error
	}{
		{
			args:      []string{},
			network:   "mainnet",
			expConfig: GetDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
		{
			args:      []string{"--sync.legacy.server", "--sync.legacy.client"},
			network:   "mainnet",
			expConfig: GetDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
		{
			args:    []string{"--sync.legacy.server", "--sync.legacy.client"},
			network: "testnet",
			expConfig: func() intelchainconfig.DnsSync {
				cfg := GetDefaultDNSSyncConfig(nodeconfig.Mainnet)
				cfg.Client = true
				cfg.Server = true
				return cfg
			}(),
		},
		{
			args:      []string{"--dns.server", "--dns.client"},
			network:   "mainnet",
			expConfig: GetDefaultDNSSyncConfig(nodeconfig.Mainnet),
		},
	}

	for i, test := range tests {
		ts := newFlagTestSuite(t, dnsSyncFlags, func(command *cobra.Command, config *intelchainconfig.IntelchainConfig) {
			config.Network.NetworkType = test.network
			applyDNSSyncFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.DNSSync, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.DNSSync, test.expConfig)
		}

		ts.tearDown()
	}
}
func TestSyncFlags(t *testing.T) {
	tests := []struct {
		args      []string
		network   string
		expConfig intelchainconfig.SyncConfig
		expErr    error
	}{
		{
			args: []string{"--sync", "--sync.downloader", "--sync.concurrency", "10", "--sync.min-peers", "10",
				"--sync.init-peers", "10", "--sync.disc.soft-low-cap", "10",
				"--sync.disc.hard-low-cap", "10", "--sync.disc.hi-cap", "10",
				"--sync.disc.batch", "10",
			},
			network: "mainnet",
			expConfig: func() intelchainconfig.SyncConfig {
				cfgSync := defaultMainnetSyncConfig
				cfgSync.Enabled = true
				cfgSync.Downloader = true
				cfgSync.StagedSync = false
				cfgSync.Concurrency = 10
				cfgSync.MinPeers = 10
				cfgSync.InitStreams = 10
				cfgSync.DiscSoftLowCap = 10
				cfgSync.DiscHardLowCap = 10
				cfgSync.DiscHighCap = 10
				cfgSync.DiscBatch = 10
				return cfgSync
			}(),
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, syncFlags, func(command *cobra.Command, config *intelchainconfig.IntelchainConfig) {
			applySyncFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.Sync, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.Sync, test.expConfig)
		}

		ts.tearDown()
	}
}

func TestShardDataFlags(t *testing.T) {
	tests := []struct {
		args      []string
		expConfig intelchainconfig.ShardDataConfig
		expErr    error
	}{
		{
			args:      []string{},
			expConfig: defaultConfig.ShardData,
		},
		{
			args: []string{"--sharddata.enable",
				"--sharddata.disk_count", "8",
				"--sharddata.shard_count", "4",
				"--sharddata.cache_time", "10",
				"--sharddata.cache_size", "512",
			},
			expConfig: intelchainconfig.ShardDataConfig{
				EnableShardData: true,
				DiskCount:       8,
				ShardCount:      4,
				CacheTime:       10,
				CacheSize:       512,
			},
		},
	}
	for i, test := range tests {
		ts := newFlagTestSuite(t, shardDataFlags, func(command *cobra.Command, config *intelchainconfig.IntelchainConfig) {
			applyShardDataFlags(command, config)
		})
		hc, err := ts.run(test.args)

		if assErr := assertError(err, test.expErr); assErr != nil {
			t.Fatalf("Test %v: %v", i, assErr)
		}
		if err != nil || test.expErr != nil {
			continue
		}
		if !reflect.DeepEqual(hc.ShardData, test.expConfig) {
			t.Errorf("Test %v:\n\t%+v\n\t%+v", i, hc.ShardData, test.expConfig)
		}

		ts.tearDown()
	}
}

type flagTestSuite struct {
	t *testing.T

	cmd *cobra.Command
	hc  intelchainconfig.IntelchainConfig
}

func newFlagTestSuite(t *testing.T, flags []cli.Flag, applyFlags func(*cobra.Command, *intelchainconfig.IntelchainConfig)) *flagTestSuite {
	cli.SetParseErrorHandle(func(err error) { t.Fatal(err) })

	ts := &flagTestSuite{hc: GetDefaultItcConfigCopy(defNetworkType)}
	ts.cmd = makeTestCommand(func(cmd *cobra.Command, args []string) {
		applyFlags(cmd, &ts.hc)
	})
	if err := cli.RegisterFlags(ts.cmd, flags); err != nil {
		t.Fatal(err)
	}

	return ts
}

func (ts *flagTestSuite) run(args []string) (intelchainconfig.IntelchainConfig, error) {
	ts.cmd.SetArgs(args)
	err := ts.cmd.Execute()
	return ts.hc, err
}

func (ts *flagTestSuite) tearDown() {
	cli.SetParseErrorHandle(func(error) {})
}

func makeTestCommand(run func(cmd *cobra.Command, args []string)) *cobra.Command {
	return &cobra.Command{
		Use: "test",
		Run: run,
	}
}

func assertError(gotErr, expErr error) error {
	if (gotErr == nil) != (expErr == nil) {
		return fmt.Errorf("error unexpected [%v] / [%v]", gotErr, expErr)
	}
	if gotErr == nil {
		return nil
	}
	if !strings.Contains(gotErr.Error(), expErr.Error()) {
		return fmt.Errorf("error unexpected [%v] / [%v]", gotErr, expErr)
	}
	return nil
}

func intPtr(i int) *int {
	return &i
}
