// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

//go:build test
// +build test

package example_chain

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testconstants "github.com/evmos/os/testutil/constants"
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

// EvmosOptionsFn defines a function type for setting app options specifically for
// the Evmos app. The function should receive the chainID and return an error if
// any.
type EvmosOptionsFn func(string) error

// NoOpEvmosOptions is a no-op function that can be used when the app does not
// need any specific configuration.
func NoOpEvmosOptions(_ string) error {
	return nil
}

// EvmosAppOptions allows to setup the global configuration
// for the Evmos chain.
func EvmosAppOptions(chainID string) error {
	// Split the revision height from the given chain ID
	id := strings.Split(chainID, "-")[0]
	coinInfo, found := ChainsCoinInfo[id]
	if !found {
		return fmt.Errorf("unknown chain id: %s", id)
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

	// sdk.RegisterDenom will automatically overwrite the base denom when the
	// new setBaseDenom() units are lower than the current base denom's units.
	return sdk.RegisterDenom(ci.Denom, math.LegacyNewDecWithPrec(1, int64(ci.Decimals)))
}
