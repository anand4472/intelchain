package network

import (
	"math/big"
	"time"

	"github.com/intelchain-itc/intelchain/consensus/engine"
	"github.com/intelchain-itc/intelchain/numeric"
)

// UtilityMetric ..
type UtilityMetric struct {
	AccumulatorSnapshot     *big.Int
	CurrentStakedPercentage numeric.Dec
	Deviation               numeric.Dec
	Adjustment              numeric.Dec
}

// NewUtilityMetricSnapshot ..
func NewUtilityMetricSnapshot(beaconchain engine.ChainReader) (*UtilityMetric, error) {
	soFarDoledOut, percentageStaked, err := WhatPercentStakedNow(
		beaconchain, time.Now().Unix(),
	)
	if err != nil {
		return nil, err
	}
	howMuchOff, adjustBy := Adjustment(*percentageStaked)
	return &UtilityMetric{
		soFarDoledOut, *percentageStaked, howMuchOff, adjustBy,
	}, nil
}