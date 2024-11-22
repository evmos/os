// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build test
// +build test

package example_chain

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/os/example_chain/eips"
	testconstants "github.com/evmos/os/testutil/constants"
	"github.com/evmos/os/x/evm/core/vm"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// ChainsCoinInfo is a map of the chain id and its corresponding EvmCoinInfo
// that allows initializing the app with different coin info based on the
// chain id
var ChainsCoinInfo = map[string]evmtypes.EvmCoinInfo{
	EighteenDecimalsChainID: {
		Denom:        ExampleChainDenom,
		DisplayDenom: ExampleChainDenom,
		Decimals:     evmtypes.EighteenDecimals,
	},
	SixDecimalsChainID: {
		Denom:        testconstants.ExampleMicroDenom,
		DisplayDenom: testconstants.ExampleDisplayDenom,
		Decimals:     evmtypes.SixDecimals,
	},
}

// InitializeAppConfiguration allows to setup the global configuration
// for tests within the Evmos EVM. We're not using the sealed flag
// and resetting the configuration to the provided one on every test setup
func InitializeAppConfiguration(chainID string) error {
	coinInfo, found := ChainsCoinInfo[chainID]
	if !found {
		// default to mainnet
		coinInfo = ChainsCoinInfo[EighteenDecimalsChainID]
	}

	// set the base denom considering if its mainnet or testnet
	if err := setBaseDenom(coinInfo); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	configurator := evmtypes.NewEVMConfigurator()
	// reset configuration to set the new one
	configurator.ResetTestConfig()
	err = configurator.
		WithExtendedEips(evmosActivators).
		WithChainConfig(ethCfg).
		WithEVMCoinInfo(baseDenom, uint8(coinInfo.Decimals)).
		Configure()
	if err != nil {
		return err
	}

	return nil
}

// EvmosActivators defines a map of opcode modifiers associated
// with a key defining the corresponding EIP.
var evmosActivators = map[string]func(*vm.JumpTable){
	"evmos_0": eips.Enable0000,
	"evmos_1": eips.Enable0001,
	"evmos_2": eips.Enable0002,
}

// setBaseDenom registers the display denom and base denom and sets the
// base denom for the chain. The function registered different values based on
// the EvmCoinInfo to allow different configurations in mainnet and testnet.
func setBaseDenom(ci evmtypes.EvmCoinInfo) (err error) {
	// Defer setting the base denom, and capture any potential error from it.
	// So when failing because the denom was already registered, we ignore it and set
	// the corresponding denom to be base denom
	defer func() {
		err = sdk.SetBaseDenom(ci.Denom)
	}()
	if err := sdk.RegisterDenom(ci.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
