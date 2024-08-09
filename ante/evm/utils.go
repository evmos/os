// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package evm

import (
	"math"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	evmtypes "github.com/evmos/evmos/v19/x/evm/types"
	anteinterfaces "github.com/evmos/os/ante/interfaces"
)

// DecoratorUtils contain a bunch of relevant variables used for a variety of checks
// throughout the verification of an Ethereum transaction.
type DecoratorUtils struct {
	EvmParams          evmtypes.Params
	EthConfig          *params.ChainConfig
	Rules              params.Rules
	Signer             ethtypes.Signer
	BaseFee            *big.Int
	EvmDenom           string
	MempoolMinGasPrice sdkmath.LegacyDec
	GlobalMinGasPrice  sdkmath.LegacyDec
	BlockTxIndex       uint64
	TxGasLimit         uint64
	GasWanted          uint64
	MinPriority        int64
	TxFee              sdk.Coins
}

// NewMonoDecoratorUtils returns a new DecoratorUtils instance.
//
// These utilities are extracted once at the beginning of the ante handle process,
// and are used throughout the entire decorator chain.
// This avoids redundant calls to the keeper and thus improves speed of transaction processing.
func NewMonoDecoratorUtils(
	ctx sdk.Context,
	ek anteinterfaces.EVMKeeper,
	fmk anteinterfaces.FeeMarketKeeper,
) (*DecoratorUtils, error) {
	evmParams := ek.GetParams(ctx)
	chainCfg := evmParams.GetChainConfig()
	ethCfg := chainCfg.EthereumConfig(ek.ChainID())
	blockHeight := big.NewInt(ctx.BlockHeight())
	rules := ethCfg.Rules(blockHeight, true)
	baseFee := ek.GetBaseFee(ctx, ethCfg)
	feeMarketParams := fmk.GetParams(ctx)

	if rules.IsLondon && baseFee == nil {
		return nil, errorsmod.Wrap(
			evmtypes.ErrInvalidBaseFee,
			"base fee is supported but evm block context value is nil",
		)
	}

	return &DecoratorUtils{
		EvmParams:          evmParams,
		EthConfig:          ethCfg,
		Rules:              rules,
		Signer:             ethtypes.MakeSigner(ethCfg, blockHeight),
		BaseFee:            baseFee,
		MempoolMinGasPrice: ctx.MinGasPrices().AmountOf(evmParams.EvmDenom),
		GlobalMinGasPrice:  feeMarketParams.MinGasPrice,
		EvmDenom:           evmParams.EvmDenom,
		BlockTxIndex:       ek.GetTxIndexTransient(ctx),
		TxGasLimit:         0,
		GasWanted:          0,
		MinPriority:        int64(math.MaxInt64),
		TxFee:              sdk.Coins{},
	}, nil
}
