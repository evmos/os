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
	anteinterfaces "github.com/evmos/os/ante/interfaces"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// DecoratorUtils contain a bunch of relevant variables used for a variety of checks
// throughout the verification of an Ethereum transaction.
type DecoratorUtils struct {
	EvmParams          evmtypes.Params
	Rules              params.Rules
	Signer             ethtypes.Signer
	BaseFee            *big.Int
	MempoolMinGasPrice sdkmath.LegacyDec
	GlobalMinGasPrice  sdkmath.LegacyDec
	BlockTxIndex       uint64
	TxGasLimit         uint64
	GasWanted          uint64
	MinPriority        int64
	TxFee              *big.Int
}

// NewMonoDecoratorUtils returns a new DecoratorUtils instance.
//
// These utilities are extracted once at the beginning of the ante handle process,
// and are used throughout the entire decorator chain.
// This avoids redundant calls to the keeper and thus improves speed of transaction processing.
//
// All prices, fees and balances are converted into 18 decimals here
// to be correctly used in the EVM.
func NewMonoDecoratorUtils(
	ctx sdk.Context,
	ek anteinterfaces.EVMKeeper,
) (*DecoratorUtils, error) {
	evmParams := ek.GetParams(ctx)
	ethCfg := evmtypes.GetChainConfig()
	evmDenom := evmtypes.GetEVMCoinDenom()
	blockHeight := big.NewInt(ctx.BlockHeight())
	rules := ethCfg.Rules(blockHeight, true)
	baseFee := ek.GetBaseFee(ctx)

	if rules.IsLondon && baseFee == nil {
		return nil, errorsmod.Wrap(
			evmtypes.ErrInvalidBaseFee,
			"base fee is supported but evm block context value is nil",
		)
	}

	globalMinGasPrice := ek.GetMinGasPrice(ctx)

	// Mempool gas price should be scaled to the 18 decimals representation.
	// If it is already a 18 decimal token, this is a no-op.
	mempoolMinGasPrice := evmtypes.ConvertAmountTo18DecimalsLegacy(ctx.MinGasPrices().AmountOf(evmDenom))

	return &DecoratorUtils{
		EvmParams:          evmParams,
		Rules:              rules,
		Signer:             ethtypes.MakeSigner(ethCfg, blockHeight),
		BaseFee:            baseFee,
		MempoolMinGasPrice: mempoolMinGasPrice,
		GlobalMinGasPrice:  globalMinGasPrice,
		BlockTxIndex:       ek.GetTxIndexTransient(ctx),
		GasWanted:          0,
		MinPriority:        int64(math.MaxInt64),
		// TxGasLimit and TxFee are set to zero because they are updated
		// summing up the values of all messages contained in a tx.
		TxGasLimit: 0,
		TxFee:      big.NewInt(0),
	}, nil
}
