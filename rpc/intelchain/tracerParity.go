package rpc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zennittians/intelchain/eth/rpc"
	"github.com/zennittians/intelchain/itc"
)

var (
	parityTraceGO = "ParityBlockTracer"
)

type PublicParityTracerService struct {
	*PublicTracerService
}

func (s *PublicParityTracerService) Transaction(ctx context.Context, hash common.Hash) (interface{}, error) {
	timer := DoMetricRPCRequest(Transaction)
	defer DoRPCRequestDuration(Transaction, timer)
	return s.TraceTransaction(ctx, hash, &itc.TraceConfig{Tracer: &parityTraceGO})
}

// trace_block RPC
func (s *PublicParityTracerService) Block(ctx context.Context, number rpc.BlockNumber) (interface{}, error) {
	timer := DoMetricRPCRequest(Block)
	defer DoRPCRequestDuration(Block, timer)

	block := s.itc.BlockChain.GetBlockByNumber(uint64(number))
	if block == nil {
		return nil, nil
	}
	if results, err := s.itc.NodeAPI.GetTraceResultByHash(block.Hash()); err == nil {
		return results, nil
	}
	results, err := s.itc.TraceBlock(ctx, block, &itc.TraceConfig{Tracer: &parityTraceGO})
	if err != nil {
		return results, err
	}
	var resultArray = make([]json.RawMessage, 0)
	for _, result := range results {
		raw, ok := result.Result.([]json.RawMessage)
		if !ok {
			return results, errors.New("tracer bug:expected []json.RawMessage")
		}
		resultArray = append(resultArray, raw...)
	}
	return resultArray, nil
}
