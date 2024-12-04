package main

import (
	"testing"

	intelchainconfig "github.com/intelchain-itc/intelchain/internal/configs/intelchain"
	"github.com/stretchr/testify/require"

	nodeconfig "github.com/intelchain-itc/intelchain/internal/configs/node"
)

var (
	V1_0_2ConfigDefault = []byte(`
Version = "1.0.2"

[BLSKeys]
  KMSConfigFile = ""
  KMSConfigSrcType = "shared"
  KMSEnabled = false
  KeyDir = "./.itc/blskeys"
  KeyFiles = []
  MaxKeys = 10
  PassEnabled = true
  PassFile = ""
  PassSrcType = "auto"
  SavePassphrase = false

[General]
  DataDir = "./"
  IsArchival = false
  IsOffline = false
  NoStaking = false
  NodeType = "validator"
  ShardID = -1

[HTTP]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9500
  RosettaEnabled = false
  RosettaPort = 9700

[Log]
  FileName = "intelchain.log"
  Folder = "./latest"
  RotateSize = 100
  Verbosity = 3

[Network]
  BootNodes = ["/ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv","/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9","/ip4/13.113.101.219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX","/ip4/99.81.170.167/tcp/12019/p2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj"]
  DNSPort = 9000
  DNSZone = "intelchain.network"
  LegacySyncing = false
  NetworkType = "mainnet"

[P2P]
  IP = "0.0.0.0"
  KeyFile = "./.itckey"
  Port = 9000

[Pprof]
  Enabled = false
  ListenAddr = "127.0.0.1:6060"

[RPCOpt]
  DebugEnabled = false
  EthRPCsEnabled = true
  StakingRPCsEnabled = true
  LegacyRPCsEnabled = true
  RpcFilterFile = "./.itc/rpc_filter.txt"

[TxPool]
  BlacklistFile = "./.itc/blacklist.txt"
  LocalAccountsFile = "./.itc/locals.txt"
  AccountQueue = 64
  GlobalQueue = 5120
  Lifetime = "30m"
  PriceBump = 1
  PriceLimit = 100e9

[WS]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9800
`)

	V1_0_3ConfigDefault = []byte(`
Version = "1.0.3"

[BLSKeys]
  KMSConfigFile = ""
  KMSConfigSrcType = "shared"
  KMSEnabled = false
  KeyDir = "./.itc/blskeys"
  KeyFiles = []
  MaxKeys = 10
  PassEnabled = true
  PassFile = ""
  PassSrcType = "auto"
  SavePassphrase = false

[General]
  DataDir = "./"
  IsArchival = false
  IsBeaconArchival = false
  IsOffline = false
  NoStaking = false
  NodeType = "validator"
  ShardID = -1

[HTTP]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9500
  RosettaEnabled = false
  RosettaPort = 9700

[Log]
  FileName = "intelchain.log"
  Folder = "./latest"
  RotateSize = 100
  Verbosity = 3

[Network]
  BootNodes = ["/ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv","/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9","/ip4/13.113.101.219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX","/ip4/99.81.170.167/tcp/12019/p2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj"]
  DNSPort = 9000
  DNSZone = "intelchain.network"
  LegacySyncing = false
  NetworkType = "mainnet"

[P2P]
  IP = "0.0.0.0"
  KeyFile = "./.itckey"
  Port = 9000

[Pprof]
  Enabled = false
  ListenAddr = "127.0.0.1:6060"

[RPCOpt]
  DebugEnabled = false
  EthRPCsEnabled = true
  StakingRPCsEnabled = true
  LegacyRPCsEnabled = true
  RpcFilterFile = "./.itc/rpc_filter.txt"

[TxPool]
  BlacklistFile = "./.itc/blacklist.txt"
  LocalAccountsFile = "./.itc/locals.txt"
  AccountQueue = 64
  GlobalQueue = 5120
  Lifetime = "30m"
  PriceBump = 1
  PriceLimit = 100e9

[WS]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9800
`)

	V1_0_4ConfigDefault = []byte(`
Version = "1.0.4"

[BLSKeys]
  KMSConfigFile = ""
  KMSConfigSrcType = "shared"
  KMSEnabled = false
  KeyDir = "./.itc/blskeys"
  KeyFiles = []
  MaxKeys = 10
  PassEnabled = true
  PassFile = ""
  PassSrcType = "auto"
  SavePassphrase = false

[General]
  DataDir = "./"
  IsArchival = false
  IsBeaconArchival = false
  IsOffline = false
  NoStaking = false
  NodeType = "validator"
  ShardID = -1

[HTTP]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9500
  RosettaEnabled = false
  RosettaPort = 9700

[Log]
  FileName = "intelchain.log"
  Folder = "./latest"
  RotateSize = 100
  Verbosity = 3

[Network]
  BootNodes = ["/dnsaddr/bootstrap.intelchain.network"]
  DNSPort = 9000
  DNSZone = "intelchain.network"
  LegacySyncing = false
  NetworkType = "mainnet"

[P2P]
  IP = "0.0.0.0"
  KeyFile = "./.itckey"
  Port = 9000

[Pprof]
  Enabled = false
  ListenAddr = "127.0.0.1:6060"

[RPCOpt]
  DebugEnabled = false
  EthRPCsEnabled = true
  StakingRPCsEnabled = true
  LegacyRPCsEnabled = true
  RpcFilterFile = "./.itc/rpc_filter.txt"

[Sync]
  Concurrency = 6
  DiscBatch = 8
  DiscHardLowCap = 6
  DiscHighCap = 128
  DiscSoftLowCap = 8
  Downloader = false
  InitStreams = 8
  LegacyClient = true
  LegacyServer = true
  MinPeers = 6

[TxPool]
  BlacklistFile = "./.itc/blacklist.txt"
  LocalAccountsFile = "./.itc/locals.txt"
  AccountQueue = 64
  GlobalQueue = 5120
  Lifetime = "30m"
  PriceBump = 1
  PriceLimit = 100e9

[WS]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9800
`)

	V1_0_4ConfigDownloaderOn = []byte(`
Version = "1.0.4"

[BLSKeys]
  KMSConfigFile = ""
  KMSConfigSrcType = "shared"
  KMSEnabled = false
  KeyDir = "./.itc/blskeys"
  KeyFiles = []
  MaxKeys = 10
  PassEnabled = true
  PassFile = ""
  PassSrcType = "auto"
  SavePassphrase = false

[General]
  DataDir = "./"
  IsArchival = false
  IsBeaconArchival = false
  IsOffline = false
  NoStaking = false
  NodeType = "validator"
  ShardID = -1

[HTTP]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9500
  RosettaEnabled = false
  RosettaPort = 9700

[Log]
  FileName = "intelchain.log"
  Folder = "./latest"
  RotateSize = 100
  Verbosity = 3

[Network]
  BootNodes = ["/dnsaddr/bootstrap.intelchain.network"]
  DNSPort = 9000
  DNSZone = "intelchain.network"
  LegacySyncing = false
  NetworkType = "mainnet"

[P2P]
  IP = "0.0.0.0"
  KeyFile = "./.itckey"
  Port = 9000

[Pprof]
  Enabled = false
  ListenAddr = "127.0.0.1:6060"

[RPCOpt]
  DebugEnabled = false
  EthRPCsEnabled = true
  StakingRPCsEnabled = true
  LegacyRPCsEnabled = true
  RpcFilterFile = "./.itc/rpc_filter.txt"

[Sync]
  Concurrency = 6
  DiscBatch = 8
  DiscHardLowCap = 6
  DiscHighCap = 128
  DiscSoftLowCap = 8
  Downloader = true
  InitStreams = 8
  LegacyClient = true
  LegacyServer = true
  MinPeers = 6

[ShardData]
  EnableShardData = false
  DiskCount = 8
  ShardCount = 4
  CacheTime = 10
  CacheSize = 512

[TxPool]
  BlacklistFile = "./.itc/blacklist.txt"
  LocalAccountsFile = "./.itc/locals.txt"
  AllowedTxsFile = "./.itc/allowedtxs.txt"
  AccountQueue = 64
  GlobalQueue = 5120
  Lifetime = "30m"
  PriceBump = 1
  PriceLimit = 100e9

[WS]
  Enabled = true
  IP = "127.0.0.1"
  Port = 9800
`)
)

func Test_migrateConf(t *testing.T) {
	defConf := getDefaultItcConfigCopy(nodeconfig.Mainnet)
	legacyDefConf := getDefaultItcConfigCopy(nodeconfig.Mainnet)
	// Versions prior to 1.0.3 use different BootNodes
	legacyDefConf.Network.BootNodes = []string{
		"/ip4/100.26.90.187/tcp/9874/p2p/Qmdfjtk6hPoyrH1zVD9PEH4zfWLo38dP2mDvvKXfh3tnEv",
		"/ip4/54.213.43.194/tcp/9874/p2p/QmZJJx6AdaoEkGLrYG4JeLCKeCKDjnFz2wfHNHxAqFSGA9",
		"/ip4/13.113.101.219/tcp/12019/p2p/QmQayinFSgMMw5cSpDUiD9pQ2WeP6WNmGxpZ6ou3mdVFJX",
		"/ip4/99.81.170.167/tcp/12019/p2p/QmRVbTpEYup8dSaURZfF6ByrMTSKa4UyUzJhSjahFzRqNj",
	}
	type args struct {
		confBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    intelchainconfig.IntelchainConfig
		wantErr bool
	}{
		{
			name: "1.0.2 to latest migration",
			args: args{
				confBytes: V1_0_2ConfigDefault,
			},
			want:    legacyDefConf,
			wantErr: false,
		},
		{
			name: "1.0.3 to latest migration",
			args: args{
				confBytes: V1_0_3ConfigDefault,
			},
			want:    legacyDefConf,
			wantErr: false,
		},
		{
			name: "1.0.4 to latest migration",
			args: args{
				confBytes: V1_0_4ConfigDefault,
			},
			want:    defConf,
			wantErr: false,
		},
		{
			name: "1.0.4 with sync downloaders on",
			args: args{
				confBytes: V1_0_4ConfigDownloaderOn,
			},
			want: func() intelchainconfig.IntelchainConfig {
				hc := defConf
				hc.Sync.Downloader = true
				hc.Sync.Enabled = true
				return hc
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		if tt.name != "1.0.4 with sync downloaders on" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := migrateConf(tt.args.confBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("migrateConf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}
