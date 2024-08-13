// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package interfaces

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/os/x/evm/core/vm"
	"github.com/evmos/os/x/evm/statedb"
	evmtypes "github.com/evmos/os/x/evm/types"
	feemarkettypes "github.com/evmos/os/x/feemarket/types"
)

// EVMKeeper exposes the required EVM keeper interface required for ante handlers
type EVMKeeper interface {
	statedb.Keeper
	DynamicFeeEVMKeeper

	NewEVM(ctx sdk.Context, msg core.Message, cfg *statedb.EVMConfig, tracer vm.EVMLogger, stateDB vm.StateDB) *vm.EVM
	DeductTxCostsFromUserBalance(ctx sdk.Context, fees sdk.Coins, from common.Address) error
	GetBalance(ctx sdk.Context, addr common.Address) *big.Int
	ResetTransientGasUsed(ctx sdk.Context)
	GetTxIndexTransient(ctx sdk.Context) uint64
	GetParams(ctx sdk.Context) evmtypes.Params
}

// FeeMarketKeeper exposes the required feemarket keeper interface required for ante handlers
type FeeMarketKeeper interface {
	GetParams(ctx sdk.Context) (params feemarkettypes.Params)
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
	GetBaseFeeEnabled(ctx sdk.Context) bool
}

// DynamicFeeEVMKeeper is a subset of EVMKeeper interface that supports dynamic fee checker
type DynamicFeeEVMKeeper interface {
	ChainID() *big.Int
	GetParams(ctx sdk.Context) evmtypes.Params
	GetBaseFee(ctx sdk.Context, ethCfg *params.ChainConfig) *big.Int
}

type ProtoTxProvider interface {
	GetProtoTx() *tx.Tx
}
