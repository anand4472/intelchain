package worker

import "github.com/intelchain-itc/intelchain/block"

type Environment interface {
	CurrentHeader() *block.Header
}
