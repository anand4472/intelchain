package common

import (
	"github.com/intelchain-itc/intelchain/consensus/quorum"
	"github.com/intelchain-itc/intelchain/crypto/bls"
	"github.com/intelchain-itc/intelchain/numeric"
)

type setRawStakeHack interface {
	SetRawStake(key bls.SerializedPublicKey, d numeric.Dec)
}

// SetRawStake is a hack, return value is if was successful or not at setting
func SetRawStake(q quorum.Decider, key bls.SerializedPublicKey, d numeric.Dec) bool {
	if setter, ok := q.(setRawStakeHack); ok {
		setter.SetRawStake(key, d)
		return true
	}
	return false
}
