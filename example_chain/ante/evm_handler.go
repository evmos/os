// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package ante

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmante "github.com/evmos/os/ante/evm"
	anteinterfaces "github.com/evmos/os/ante/interfaces"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// newMonoEVMAnteHandler creates the sdk.AnteHandler implementation for the EVM transactions.
func newMonoEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		NewEVMMonoDecorator(
			options.AccountKeeper,
			options.BankKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.DistributionKeeper,
			options.StakingKeeper,
			options.MaxTxGasWanted,
		),
	)
}

// MonoDecorator is a single decorator that handles all the prechecks for
// ethereum transactions.
type MonoDecorator struct {
	accountKeeper      evmtypes.AccountKeeper
	bankKeeper         evmtypes.BankKeeper
	feeMarketKeeper    anteinterfaces.FeeMarketKeeper
	evmKeeper          anteinterfaces.EVMKeeper
	distributionKeeper anteinterfaces.DistributionKeeper
	stakingKeeper      anteinterfaces.StakingKeeper
	maxGasWanted       uint64
}

// NewEVMMonoDecorator creates the 'mono' decorator, that is used to run the ante handle logic
// for EVM transactions on the chain.
//
// TODO: make partners not have to duplicate this in their repo
func NewEVMMonoDecorator(
	accountKeeper evmtypes.AccountKeeper,
	bankKeeper evmtypes.BankKeeper,
	feeMarketKeeper anteinterfaces.FeeMarketKeeper,
	evmKeeper anteinterfaces.EVMKeeper,
	distributionKeeper anteinterfaces.DistributionKeeper,
	stakingKeeper anteinterfaces.StakingKeeper,
	maxGasWanted uint64,
) MonoDecorator {
	return MonoDecorator{
		accountKeeper:      accountKeeper,
		bankKeeper:         bankKeeper,
		feeMarketKeeper:    feeMarketKeeper,
		evmKeeper:          evmKeeper,
		distributionKeeper: distributionKeeper,
		stakingKeeper:      stakingKeeper,
		maxGasWanted:       maxGasWanted,
	}
}

// AnteHandle handles the entire decorator chain using a mono decorator.
func (md MonoDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// 0. Basic validation of the transaction
	var txFeeInfo *txtypes.Fee
	if !ctx.IsReCheckTx() {
		txFeeInfo, err = evmante.ValidateTx(tx)
		if err != nil {
			return ctx, err
		}
	}

	// 1. setup ctx
	ctx, err = evmante.SetupContext(ctx, tx, md.evmKeeper)
	if err != nil {
		return ctx, err
	}

	// 2. get utils
	decUtils, err := evmante.NewMonoDecoratorUtils(ctx, md.evmKeeper, md.feeMarketKeeper)
	if err != nil {
		return ctx, err
	}

	// Use the lowest priority of all the messages as the final one.
	for i, msg := range tx.GetMsgs() {
		ethMsg, txData, from, err := evmtypes.UnpackEthMsg(msg)
		if err != nil {
			return ctx, err
		}

		feeAmt := txData.Fee()
		gas := txData.GetGas()
		fee := sdkmath.LegacyNewDecFromBigInt(feeAmt)
		gasLimit := sdkmath.LegacyNewDecFromBigInt(new(big.Int).SetUint64(gas))

		// 2. mempool inclusion fee
		if ctx.IsCheckTx() && !simulate {
			if err := evmante.CheckMempoolFee(fee, decUtils.MempoolMinGasPrice, gasLimit, decUtils.Rules.IsLondon); err != nil {
				return ctx, err
			}
		}

		// 3. min gas price (global min fee)
		if txData.TxType() == ethtypes.DynamicFeeTxType && decUtils.BaseFee != nil {
			feeAmt = txData.EffectiveFee(decUtils.BaseFee)
			fee = sdkmath.LegacyNewDecFromBigInt(feeAmt)
		}

		if err := evmante.CheckGlobalFee(fee, decUtils.GlobalMinGasPrice, gasLimit); err != nil {
			return ctx, err
		}

		// 4. validate msg contents
		err = evmante.ValidateMsg(
			decUtils.EvmParams,
			txData,
			from,
		)
		if err != nil {
			return ctx, err
		}

		// 5. signature verification
		if err := evmante.SignatureVerification(
			ethMsg,
			decUtils.Signer,
			decUtils.EvmParams.AllowUnprotectedTxs,
		); err != nil {
			return ctx, err
		}

		// NOTE: sender address has been verified and cached
		from = ethMsg.GetFrom()

		// 6. account balance verification
		fromAddr := common.HexToAddress(ethMsg.From)
		// TODO: Use account from AccountKeeper instead
		account := md.evmKeeper.GetAccount(ctx, fromAddr)
		if err := evmante.VerifyAccountBalance(
			ctx,
			md.accountKeeper,
			account,
			fromAddr,
			txData,
		); err != nil {
			return ctx, err
		}

		// 7. can transfer
		coreMsg, err := ethMsg.AsMessage(decUtils.Signer, decUtils.BaseFee)
		if err != nil {
			return ctx, errorsmod.Wrapf(
				err,
				"failed to create an ethereum core.Message from signer %T", decUtils.Signer,
			)
		}

		if err := evmante.CanTransfer(
			ctx,
			md.evmKeeper,
			coreMsg,
			decUtils.BaseFee,
			decUtils.EthConfig,
			decUtils.EvmParams,
			decUtils.Rules.IsLondon,
		); err != nil {
			return ctx, err
		}

		// 8. gas consumption
		msgFees, err := evmkeeper.VerifyFee(
			txData,
			decUtils.EvmDenom,
			decUtils.BaseFee,
			decUtils.Rules.IsHomestead,
			decUtils.Rules.IsIstanbul,
			ctx.IsCheckTx(),
		)
		if err != nil {
			return ctx, err
		}

		err = evmante.ConsumeFeesAndEmitEvent(
			ctx,
			&evmante.ConsumeGasKeepers{
				Bank:         md.bankKeeper,
				Distribution: md.distributionKeeper,
				Evm:          md.evmKeeper,
				Staking:      md.stakingKeeper,
			},
			msgFees,
			from,
		)
		if err != nil {
			return ctx, err
		}

		gasWanted := evmante.UpdateCumulativeGasWanted(
			ctx,
			txData.GetGas(),
			md.maxGasWanted,
			decUtils.GasWanted,
		)
		decUtils.GasWanted = gasWanted

		minPriority := evmante.GetMsgPriority(
			txData,
			decUtils.MinPriority,
			decUtils.BaseFee,
		)
		decUtils.MinPriority = minPriority

		txFee := evmante.UpdateCumulativeTxFee(
			decUtils.TxFee,
			txData.Fee(),
			decUtils.EvmDenom,
		)
		decUtils.TxFee = txFee
		decUtils.TxGasLimit += gas

		// 10. increment sequence
		acc := md.accountKeeper.GetAccount(ctx, from)
		if acc == nil {
			// safety check: shouldn't happen
			return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress,
				"account %s does not exist", acc)
		}

		if err := evmante.IncrementNonce(ctx, md.accountKeeper, acc, txData.GetNonce()); err != nil {
			return ctx, err
		}

		// 11. gas wanted
		if err := evmante.CheckGasWanted(ctx, md.feeMarketKeeper, tx, decUtils.Rules.IsLondon); err != nil {
			return ctx, err
		}

		// 12. emit events
		txIdx := uint64(i) // nosec: G115
		evmante.EmitTxHashEvent(ctx, ethMsg, decUtils.BlockTxIndex, txIdx)
	}

	if err := evmante.CheckTxFee(txFeeInfo, decUtils.TxFee, decUtils.TxGasLimit); err != nil {
		return ctx, err
	}

	ctx, err = evmante.CheckBlockGasLimit(ctx, decUtils.GasWanted, decUtils.MinPriority)
	if err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}
