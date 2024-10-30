package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/zennittians/intelchain/common/denominations"
	itcTypes "github.com/zennittians/intelchain/core/types"
	"github.com/zennittians/intelchain/itc"
	"github.com/zennittians/intelchain/rosetta/common"
	stakingTypes "github.com/zennittians/intelchain/staking/types"
)

const (
	// DefaultGasPrice ..
	DefaultGasPrice = 100 * denominations.Intello
)

// ConstructAPI implements the server.ConstructAPIServicer interface.
type ConstructAPI struct {
	itc            *itc.Intelchain
	signer         itcTypes.Signer
	stakingSigner  stakingTypes.Signer
	evmCallTimeout time.Duration
}

// NewConstructionAPI creates a new instance of a ConstructAPI.
func NewConstructionAPI(itc *itc.Intelchain) server.ConstructionAPIServicer {
	return &ConstructAPI{
		itc:            itc,
		signer:         itcTypes.NewEIP155Signer(new(big.Int).SetUint64(itc.ChainID)),
		stakingSigner:  stakingTypes.NewEIP155Signer(new(big.Int).SetUint64(itc.ChainID)),
		evmCallTimeout: itc.NodeAPI.GetConfig().NodeConfig.RPCServer.EvmCallTimeout,
	}
}

// ConstructionDerive implements the /construction/derive endpoint.
func (s *ConstructAPI) ConstructionDerive(
	ctx context.Context, request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier, s.itc.ShardID); err != nil {
		return nil, err
	}
	address, rosettaError := getAddressFromPublicKey(request.PublicKey)
	if rosettaError != nil {
		return nil, rosettaError
	}
	accountID, rosettaError := newAccountIdentifier(*address)
	if rosettaError != nil {
		return nil, rosettaError
	}
	return &types.ConstructionDeriveResponse{
		AccountIdentifier: accountID,
	}, nil
}

// getAddressFromPublicKey assumes that data is a compressed secp256k1 public key
func getAddressFromPublicKey(
	key *types.PublicKey,
) (*ethCommon.Address, *types.Error) {
	if key == nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": "nil key",
		})
	}
	if key.CurveType != common.CurveType {
		return nil, common.NewError(common.UnsupportedCurveTypeError, map[string]interface{}{
			"message": fmt.Sprintf("currently only support %v", common.CurveType),
		})
	}
	// underlying eth crypto lib uses secp256k1
	publicKey, err := crypto.DecompressPubkey(key.Bytes)
	if err != nil {
		return nil, common.NewError(common.CatchAllError, map[string]interface{}{
			"message": err.Error(),
		})
	}
	address := crypto.PubkeyToAddress(*publicKey)
	return &address, nil
}
