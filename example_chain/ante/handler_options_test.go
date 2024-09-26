package ante_test

import (
	"testing"

	"github.com/evmos/os/ante"
	ethante "github.com/evmos/os/ante/evm"
	chainante "github.com/evmos/os/example_chain/ante"
	"github.com/evmos/os/testutil/integration/os/network"
	"github.com/evmos/os/types"
	"github.com/stretchr/testify/require"
)

func TestValidateHandlerOptions(t *testing.T) {
	nw := network.NewUnitTestNetwork()
	cases := []struct {
		name    string
		options chainante.HandlerOptions
		expPass bool
	}{
		{
			"fail - empty options",
			chainante.HandlerOptions{},
			false,
		},
		{
			"fail - empty account keeper",
			chainante.HandlerOptions{
				Cdc:           nw.App.AppCodec(),
				AccountKeeper: nil,
			},
			false,
		},
		{
			"fail - empty bank keeper",
			chainante.HandlerOptions{
				Cdc:           nw.App.AppCodec(),
				AccountKeeper: nw.App.AccountKeeper,
				BankKeeper:    nil,
			},
			false,
		},
		{
			"fail - empty distribution keeper",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nil,

				IBCKeeper: nil,
			},
			false,
		},
		{
			"fail - empty IBC keeper",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,

				IBCKeeper: nil,
			},
			false,
		},
		{
			"fail - empty staking keeper",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,

				IBCKeeper:     nw.App.IBCKeeper,
				StakingKeeper: nil,
			},
			false,
		},
		{
			"fail - empty fee market keeper",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,

				IBCKeeper:       nw.App.IBCKeeper,
				StakingKeeper:   nw.App.StakingKeeper,
				FeeMarketKeeper: nil,
			},
			false,
		},
		{
			"fail - empty EVM keeper",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,
				IBCKeeper:          nw.App.IBCKeeper,
				StakingKeeper:      nw.App.StakingKeeper,
				FeeMarketKeeper:    nw.App.FeeMarketKeeper,
				EvmKeeper:          nil,
			},
			false,
		},
		{
			"fail - empty signature gas consumer",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,
				IBCKeeper:          nw.App.IBCKeeper,
				StakingKeeper:      nw.App.StakingKeeper,
				FeeMarketKeeper:    nw.App.FeeMarketKeeper,
				EvmKeeper:          nw.App.EVMKeeper,
				SigGasConsumer:     nil,
			},
			false,
		},
		{
			"fail - empty signature mode handler",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,
				IBCKeeper:          nw.App.IBCKeeper,
				StakingKeeper:      nw.App.StakingKeeper,
				FeeMarketKeeper:    nw.App.FeeMarketKeeper,
				EvmKeeper:          nw.App.EVMKeeper,
				SigGasConsumer:     ante.SigVerificationGasConsumer,
				SignModeHandler:    nil,
			},
			false,
		},
		{
			"fail - empty tx fee checker",
			chainante.HandlerOptions{
				Cdc:                nw.App.AppCodec(),
				AccountKeeper:      nw.App.AccountKeeper,
				BankKeeper:         nw.App.BankKeeper,
				DistributionKeeper: nw.App.DistrKeeper,
				IBCKeeper:          nw.App.IBCKeeper,
				StakingKeeper:      nw.App.StakingKeeper,
				FeeMarketKeeper:    nw.App.FeeMarketKeeper,
				EvmKeeper:          nw.App.EVMKeeper,
				SigGasConsumer:     ante.SigVerificationGasConsumer,
				SignModeHandler:    nw.App.GetTxConfig().SignModeHandler(),
				TxFeeChecker:       nil,
			},
			false,
		},
		{
			"success - default app options",
			chainante.HandlerOptions{
				Cdc:                    nw.App.AppCodec(),
				AccountKeeper:          nw.App.AccountKeeper,
				BankKeeper:             nw.App.BankKeeper,
				DistributionKeeper:     nw.App.DistrKeeper,
				ExtensionOptionChecker: types.HasDynamicFeeExtensionOption,
				EvmKeeper:              nw.App.EVMKeeper,
				StakingKeeper:          nw.App.StakingKeeper,
				FeegrantKeeper:         nw.App.FeeGrantKeeper,
				IBCKeeper:              nw.App.IBCKeeper,
				FeeMarketKeeper:        nw.App.FeeMarketKeeper,
				SignModeHandler:        nw.GetEncodingConfig().TxConfig.SignModeHandler(),
				SigGasConsumer:         ante.SigVerificationGasConsumer,
				MaxTxGasWanted:         40000000,
				TxFeeChecker:           ethante.NewDynamicFeeChecker(nw.App.EVMKeeper),
			},
			true,
		},
	}

	for _, tc := range cases {
		err := tc.options.Validate()
		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
