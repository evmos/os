// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package example_chain

import (
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/evmos/os/cmd/config"
	chainconfig "github.com/evmos/os/example_chain/osd/config"
	feemarkettypes "github.com/evmos/os/x/feemarket/types"
	"github.com/stretchr/testify/require"
)

// SetupOptions defines arguments that are passed into `Simapp` constructor.
type SetupOptions struct {
	Logger  log.Logger
	DB      *dbm.MemDB
	AppOpts servertypes.AppOptions
}

func init() {
	// we're setting the minimum gas price to 0 to simplify the tests
	feemarkettypes.DefaultMinGasPrice = math.LegacyZeroDec()

	// Set the global SDK config for the tests
	cfg := sdk.GetConfig()
	chainconfig.SetBech32Prefixes(cfg)
	config.SetBip44CoinType(cfg)
}

func setup(withGenesis bool, invCheckPeriod uint, chainID string) (*ExampleChain, GenesisState) {
	db := dbm.NewMemDB()

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = DefaultNodeHome
	appOptions[server.FlagInvCheckPeriod] = invCheckPeriod

	app := NewExampleApp(log.NewNopLogger(), db, nil, true, appOptions, EvmosAppOptions, baseapp.SetChainID(chainID))
	if withGenesis {
		return app, app.DefaultGenesis()
	}

	return app, GenesisState{}
}

// Setup initializes a new ExampleChain. A Nop logger is set in ExampleChain.
func Setup(t *testing.T, chainID string) *ExampleChain {
	t.Helper()

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000000000000))),
	}

	app := SetupWithGenesisValSet(t, chainID, valSet, []authtypes.GenesisAccount{acc}, balance)

	return app
}

// SetupWithGenesisValSet initializes a new ExampleChain with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the simapp from first genesis
// account. A Nop logger is set in ExampleChain.
func SetupWithGenesisValSet(t *testing.T, chainID string, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) *ExampleChain {
	t.Helper()

	app, genesisState := setup(true, 5, chainID)
	genesisState, err := simtestutil.GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
	require.NoError(t, err)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	if _, err = app.InitChain(
		&abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: simtestutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
			ChainId:         chainID,
		},
	); err != nil {
		panic(fmt.Sprintf("app.InitChain failed: %v", err))
	}

	// NOTE: we are NOT committing the changes here as opposed to the function from simapp
	// because that would already adjust e.g. the base fee in the params.
	// We want to keep the genesis state as is for the tests unless we commit the changes manually.

	return app
}

// SetupTestingApp initializes the IBC-go testing application
// need to keep this design to comply with the ibctesting SetupTestingApp func
// and be able to set the chainID for the tests properly
func SetupTestingApp(chainID string) func() (ibctesting.TestingApp, map[string]json.RawMessage) {
	return func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		db := dbm.NewMemDB()
		app := NewExampleApp(
			log.NewNopLogger(),
			db, nil, true,
			simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
			EvmosAppOptions,
			baseapp.SetChainID(chainID),
		)
		return app, app.DefaultGenesis()
	}
}
