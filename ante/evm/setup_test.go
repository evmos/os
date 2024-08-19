package evm_test

import (
	"math"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/client"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/core/types"
	evmosante "github.com/evmos/os/ante"
	"github.com/evmos/os/encoding"
	"github.com/evmos/os/ethereum/eip712"
	example_app "github.com/evmos/os/example_chain"
	chainante "github.com/evmos/os/example_chain/ante"
	chainutil "github.com/evmos/os/example_chain/testutil"
	"github.com/evmos/os/testutil"
	evmtypes "github.com/evmos/os/x/evm/types"
	feemarkettypes "github.com/evmos/os/x/feemarket/types"
)

type AnteTestSuite struct {
	suite.Suite

	ctx                      sdk.Context
	app                      *example_app.ExampleChain
	clientCtx                client.Context
	anteHandler              sdk.AnteHandler
	ethSigner                types.Signer
	enableFeemarket          bool
	enableLondonHF           bool
	evmParamsOption          func(*evmtypes.Params)
	useLegacyEIP712TypedData bool
}

const TestGasLimit uint64 = 100000

func (suite *AnteTestSuite) SetupTest() {
	checkTx := false

	suite.app = chainutil.EthSetup(checkTx, testutil.ExampleChainID, func(app *example_app.ExampleChain, genesis simapp.GenesisState) simapp.GenesisState {
		if suite.enableFeemarket {
			// setup feemarketGenesis params
			feemarketGenesis := feemarkettypes.DefaultGenesisState()
			feemarketGenesis.Params.EnableHeight = 1
			feemarketGenesis.Params.NoBaseFee = false
			// Verify feeMarket genesis
			err := feemarketGenesis.Validate()
			suite.Require().NoError(err)
			genesis[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(feemarketGenesis)
		}
		evmGenesis := evmtypes.DefaultGenesisState()
		evmGenesis.Params.EvmDenom = example_app.ExampleChainDenom
		evmGenesis.Params.AllowUnprotectedTxs = false
		if !suite.enableLondonHF {
			maxInt := sdkmath.NewInt(math.MaxInt64)
			evmGenesis.Params.ChainConfig.LondonBlock = &maxInt
			evmGenesis.Params.ChainConfig.ArrowGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.GrayGlacierBlock = &maxInt
			evmGenesis.Params.ChainConfig.MergeNetsplitBlock = &maxInt
			evmGenesis.Params.ChainConfig.ShanghaiBlock = &maxInt
			evmGenesis.Params.ChainConfig.CancunBlock = &maxInt
		}
		if suite.evmParamsOption != nil {
			suite.evmParamsOption(&evmGenesis.Params)
		}
		genesis[evmtypes.ModuleName] = app.AppCodec().MustMarshalJSON(evmGenesis)
		return genesis
	})

	suite.ctx = suite.app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 2, ChainID: testutil.ExampleChainID, Time: time.Now().UTC()})
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(testutil.ExampleAttoDenom, sdkmath.OneInt())))
	suite.ctx = suite.ctx.WithBlockGasMeter(storetypes.NewGasMeter(1000000000000000000))

	// set staking denomination to Evmos denom
	params := suite.app.StakingKeeper.GetParams(suite.ctx)
	params.BondDenom = example_app.ExampleChainDenom
	err := suite.app.StakingKeeper.SetParams(suite.ctx, params)
	suite.Require().NoError(err)

	infCtx := suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	err = suite.app.AccountKeeper.SetParams(infCtx, authtypes.DefaultParams())
	suite.Require().NoError(err)

	encodingConfig := encoding.MakeConfig(example_app.ModuleBasics)
	// We're using TestMsg amino encoding in some tests, so register it here.
	encodingConfig.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)
	eip712.SetEncodingConfig(encodingConfig)

	suite.clientCtx = client.Context{}.WithTxConfig(encodingConfig.TxConfig)

	suite.Require().NotNil(suite.app.AppCodec())

	anteHandler := chainante.NewAnteHandler(chainante.HandlerOptions{
		Cdc:                suite.app.AppCodec(),
		AccountKeeper:      suite.app.AccountKeeper,
		BankKeeper:         suite.app.BankKeeper,
		DistributionKeeper: suite.app.DistrKeeper,
		EvmKeeper:          suite.app.EVMKeeper,
		FeegrantKeeper:     suite.app.FeeGrantKeeper,
		StakingKeeper:      suite.app.StakingKeeper,
		FeeMarketKeeper:    suite.app.FeeMarketKeeper,
		SignModeHandler:    encodingConfig.TxConfig.SignModeHandler(),
		SigGasConsumer:     evmosante.SigVerificationGasConsumer,
	})

	suite.anteHandler = anteHandler
	suite.ethSigner = types.LatestSignerForChainID(suite.app.EVMKeeper.ChainID())
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{
		enableLondonHF: true,
	})

	// Re-run the tests with EIP-712 Legacy encodings to ensure backwards compatibility.
	// LegacyEIP712Extension should not be run with current TypedData encodings, since they are not compatible.
	suite.Run(t, &AnteTestSuite{
		enableLondonHF:           true,
		useLegacyEIP712TypedData: true,
	})

	suite.Run(t, &AnteTestSuite{
		enableLondonHF:           true,
		useLegacyEIP712TypedData: true,
	})
}
