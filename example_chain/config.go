//go:build !test
// +build !test

package example_chain

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	chainconfig "github.com/evmos/os/example_chain/osd/config"
	evmtypes "github.com/evmos/os/x/evm/types"
)

var sealed = false

// InitializeAppConfiguration allows to setup the global configuration
// for the evmOS EVM.
func InitializeAppConfiguration(chainID string) error {
	if sealed {
		return nil
	}

	// When calling any CLI command, it creates a tempApp inside RootCmdHandler with an empty chainID.
	if chainID == "" {
		return nil
	}

	// set the denom info for the chain
	if err := setBaseDenomWithChainID(); err != nil {
		return err
	}

	baseDenom, err := sdk.GetBaseDenom()
	if err != nil {
		return err
	}

	ethCfg := evmtypes.DefaultChainConfig(chainID)

	err = evmtypes.NewEVMConfigurator().
		WithChainConfig(ethCfg).
		// NOTE: we're using the 18 decimals default for the example chain
		WithEVMCoinInfo(baseDenom, uint8(evmtypes.EighteenDecimals)).
		Configure()
	if err != nil {
		return err
	}

	sealed = true
	return nil
}

// setBaseDenomWithChainID registers the display denom and base denom and sets the
// base denom for the chain.
func setBaseDenomWithChainID() error {
	if err := sdk.RegisterDenom(chainconfig.DisplayDenom, math.LegacyOneDec()); err != nil {
		return err
	}

	// sdk.RegisterDenom will automatically overwrite the base denom when the new denom units are lower than the current base denom's units.
	return sdk.RegisterDenom(chainconfig.BaseDenom, math.LegacyNewDecWithPrec(1, chainconfig.BaseDenomUnit))
}
