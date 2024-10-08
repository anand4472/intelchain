package blockproposal

import (
	msg_pb "github.com/zennittians/intelchain/api/proto/message"
	"github.com/zennittians/intelchain/consensus"
	"github.com/zennittians/intelchain/internal/utils"
)

// Service is a block proposal service.
type Service struct {
	stopChan    chan struct{}
	stoppedChan chan struct{}
	c           *consensus.Consensus
	messageChan chan *msg_pb.Message
}

// New returns a block proposal service.
func New(c *consensus.Consensus) *Service {
	return &Service{
		c:           c,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// Start starts block proposal service.
func (s *Service) Start() error {
	s.run()
	return nil
}

func (s *Service) run() {
	s.c.WaitForConsensusReadyV2(s.stopChan, s.stoppedChan)
}

// Stop stops block proposal service.
func (s *Service) Stop() error {
	utils.Logger().Info().Msg("Stopping block proposal service.")
	s.stopChan <- struct{}{}
	<-s.stoppedChan
	utils.Logger().Info().Msg("Role conversion stopped.")
	return nil
}
