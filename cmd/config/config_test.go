package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	intelchainconfig "github.com/zennittians/intelchain/internal/configs/intelchain"

	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
)

type testCfgOpt func(config *intelchainconfig.IntelchainConfig)

func makeTestConfig(nt nodeconfig.NetworkType, opt testCfgOpt) intelchainconfig.IntelchainConfig {
	cfg := GetDefaultItcConfigCopy(nt)
	if opt != nil {
		opt(&cfg)
	}
	return cfg
}

var testBaseDir = ".testdata"

func init() {
	if _, err := os.Stat(testBaseDir); os.IsNotExist(err) {
		os.MkdirAll(testBaseDir, 0777)
	}
}

func TestV1_0_4Config(t *testing.T) {
	testConfig := `
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
  NoStaking = false
  NodeType = "validator"
  ShardID = -1

[HTTP]
  Enabled = true
  IP = "146.190.38.146"
  Port = 9500

[Log]
  Console = false
  FileName = "intelchain.log"
  Folder = "./latest"
  RotateSize = 100
  RotateCount = 0
  RotateMaxAge = 0
  Verbosity = 3

[Network]
  BootNodes = ["/dnsaddr/bootstrap.t.intelchain.org"]
  DNSPort = 9000
  DNSZone = "t.intelchain.org"
  LegacySyncing = false
  NetworkType = "mainnet"

[P2P]
  KeyFile = "./.itckey"
  Port = 9000

[Pprof]
  Enabled = false
  ListenAddr = "146.190.38.146:6060"

[TxPool]
  BlacklistFile = "./.itc/blacklist.txt"
  LocalAccountsFile = "./.itc/locals.txt"
  AllowedTxsFile = "./.itc/allowedtxs.txt"
  AccountQueue = 64
  GlobalQueue = 5120
  Lifetime = "30m"
  PriceBump = 1
  PriceLimit = 100e9

[Sync]
  Downloader = false
  Concurrency = 6
  DiscBatch = 8
  DiscHardLowCap = 6
  DiscHighCap = 128
  DiscSoftLowCap = 8
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

[WS]
  Enabled = true
  IP = "146.190.38.146"
  Port = 9800`
	testDir := filepath.Join(testBaseDir, t.Name())
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)
	file := filepath.Join(testDir, "test.config")
	err := os.WriteFile(file, []byte(testConfig), 0644)
	if err != nil {
		t.Fatal(err)
	}
	config, migratedFrom, err := loadIntelchainConfig(file)
	if err != nil {
		t.Fatal(err)
	}
	defConf := GetDefaultItcConfigCopy(nodeconfig.Mainnet)
	if config.HTTP.RosettaEnabled {
		t.Errorf("Expected rosetta http server to be disabled when loading old config")
	}
	if config.General.IsOffline {
		t.Errorf("Expect node to de online when loading old config")
	}
	if config.P2P.IP != defConf.P2P.IP {
		t.Errorf("Expect default p2p IP if old config is provided")
	}
	if migratedFrom != "1.0.4" {
		t.Errorf("Expected config version: 1.0.4, not %v", config.Version)
	}
	config.Version = defConf.Version // Shortcut for testing, value checked above
	require.Equal(t, config, defConf)
}

func TestPersistConfig(t *testing.T) {
	testDir := filepath.Join(testBaseDir, t.Name())
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0777)

	tests := []struct {
		config intelchainconfig.IntelchainConfig
	}{
		{
			config: makeTestConfig("mainnet", nil),
		},
		{
			config: makeTestConfig("devnet", nil),
		},
		{
			config: makeTestConfig("mainnet", func(cfg *intelchainconfig.IntelchainConfig) {
				consensus := GetDefaultConsensusConfigCopy()
				cfg.Consensus = &consensus

				devnet := GetDefaultDevnetConfigCopy()
				cfg.Devnet = &devnet

				revert := GetDefaultRevertConfigCopy()
				cfg.Revert = &revert

				webHook := "web hook"
				cfg.Legacy = &intelchainconfig.LegacyConfig{
					WebHookConfig:         &webHook,
					TPBroadcastInvalidTxn: &trueBool,
				}

				logCtx := GetDefaultLogContextCopy()
				cfg.Log.Context = &logCtx
			}),
		},
	}
	for i, test := range tests {
		file := filepath.Join(testDir, fmt.Sprintf("%d.conf", i))

		if err := writeIntelchainConfigToFile(test.config, file); err != nil {
			t.Fatal(err)
		}
		config, _, err := loadIntelchainConfig(file)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(config, test.config) {
			t.Errorf("Test %v: unexpected config \n\t%+v \n\t%+v", i, config, test.config)
		}
	}
}
