package bootnode

import (
	"testing"

	intelchainConfigs "github.com/zennittians/intelchain/cmd/config"
	"github.com/zennittians/intelchain/crypto/bls"
	nodeconfig "github.com/zennittians/intelchain/internal/configs/node"
	"github.com/zennittians/intelchain/internal/utils"
	"github.com/zennittians/intelchain/p2p"
)

func TestNewBootNode(t *testing.T) {
	blsKey := bls.RandPrivateKey()
	pubKey := blsKey.GetPublicKey()
	leader := p2p.Peer{IP: "127.0.0.1", Port: "8882", ConsensusPubKey: pubKey}
	priKey, _, _ := utils.GenKeyP2P("127.0.0.1", "9902")
	host, err := p2p.NewHost(p2p.HostConfig{
		Self:      &leader,
		BLSKey:    priKey,
		BootNodes: nil,
	})
	if err != nil {
		t.Fatalf("new boot node host failure: %v", err)
	}

	hc := intelchainConfigs.GetDefaultItcConfigCopy(nodeconfig.NetworkType(nodeconfig.Devnet))
	node := New(host, &hc)

	if node == nil {
		t.Error("boot node creation failed")
	}
}
