// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package network

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/evmos/os/encoding"
	exampleapp "github.com/evmos/os/example_chain"
	erc20types "github.com/evmos/os/x/erc20/types"
	evmtypes "github.com/evmos/os/x/evm/types"
	feemarkettypes "github.com/evmos/os/x/feemarket/types"
)

func getQueryHelper(ctx sdktypes.Context) *baseapp.QueryServiceTestHelper {
	encCfg := encoding.MakeConfig(exampleapp.ModuleBasics)
	interfaceRegistry := encCfg.InterfaceRegistry
	// This is needed so that state changes are not committed in precompiles
	// simulations.
	cacheCtx, _ := ctx.CacheContext()
	return baseapp.NewQueryServerTestHelper(cacheCtx, interfaceRegistry)
}

// TODO: remove Evmos specific stuff
func (n *IntegrationNetwork) GetERC20Client() erc20types.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	erc20types.RegisterQueryServer(queryHelper, n.app.Erc20Keeper)
	return erc20types.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetEvmClient() evmtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	evmtypes.RegisterQueryServer(queryHelper, n.app.EVMKeeper)
	return evmtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetGovClient() govtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	govtypes.RegisterQueryServer(queryHelper, n.app.GovKeeper)
	return govtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetBankClient() banktypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	banktypes.RegisterQueryServer(queryHelper, n.app.BankKeeper)
	return banktypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetFeeMarketClient() feemarkettypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	feemarkettypes.RegisterQueryServer(queryHelper, n.app.FeeMarketKeeper)
	return feemarkettypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetAuthClient() authtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	authtypes.RegisterQueryServer(queryHelper, n.app.AccountKeeper)
	return authtypes.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetAuthzClient() authz.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	authz.RegisterQueryServer(queryHelper, n.app.AuthzKeeper)
	return authz.NewQueryClient(queryHelper)
}

func (n *IntegrationNetwork) GetStakingClient() stakingtypes.QueryClient {
	queryHelper := getQueryHelper(n.GetContext())
	stakingtypes.RegisterQueryServer(queryHelper, stakingkeeper.Querier{Keeper: n.app.StakingKeeper})
	return stakingtypes.NewQueryClient(queryHelper)
}
