package p2ptests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	p2ptypes "github.com/zennittians/intelchain/p2p/types"
	"github.com/zennittians/intelchain/test/helpers"
)

func TestMultiAddressParsing(t *testing.T) {
	t.Parallel()

	multiAddresses, err := p2ptypes.StringsToMultiAddrs(helpers.Bootnodes)
	assert.NoError(t, err)
	assert.Equal(t, len(helpers.Bootnodes), len(multiAddresses))

	for index, multiAddress := range multiAddresses {
		assert.Equal(t, multiAddress.String(), helpers.Bootnodes[index])
	}
}

func TestAddressListConversionToString(t *testing.T) {
	t.Parallel()

	multiAddresses, err := p2ptypes.StringsToMultiAddrs(helpers.Bootnodes)
	assert.NoError(t, err)
	assert.Equal(t, len(helpers.Bootnodes), len(multiAddresses))

	expected := strings.Join(helpers.Bootnodes[:], ",")
	var addressList p2ptypes.AddrList = multiAddresses
	assert.Equal(t, expected, addressList.String())
}

func TestAddressListConversionFromString(t *testing.T) {
	t.Parallel()

	multiAddresses, err := p2ptypes.StringsToMultiAddrs(helpers.Bootnodes)
	assert.NoError(t, err)
	assert.Equal(t, len(helpers.Bootnodes), len(multiAddresses))

	addressString := strings.Join(helpers.Bootnodes[:], ",")
	var addressList p2ptypes.AddrList = multiAddresses
	addressList.Set(addressString)
	assert.Equal(t, len(addressList), len(multiAddresses))
	assert.Equal(t, addressList[0], multiAddresses[0])
}
