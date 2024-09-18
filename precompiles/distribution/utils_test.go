package distribution_test

import (
	"encoding/json"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	sdkstaking "github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	exampleapp "github.com/evmos/os/example_chain"
	chainutil "github.com/evmos/os/example_chain/testutil"
	"github.com/evmos/os/precompiles/distribution"
	evmosutil "github.com/evmos/os/testutil"
	evmosutiltx "github.com/evmos/os/testutil/tx"
	evmostypes "github.com/evmos/os/types"
	"github.com/evmos/os/x/evm/statedb"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// SetupWithGenesisValSet initializes a new evmOS application with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit (10^6) in the default token of the simapp from first genesis
// account. A Nop logger is set in SimApp.
func (s *PrecompileTestSuite) SetupWithGenesisValSet(valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, balances ...banktypes.Balance) {
	appI, genesisState := exampleapp.SetupTestingApp(evmosutil.ExampleChainID)()
	app, ok := appI.(*exampleapp.ExampleChain)
	s.Require().True(ok)

	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.TokensFromConsensusPower(1, evmostypes.AttoPowerReduction)

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		s.Require().NoError(err)
		pkAny, err := codectypes.NewAnyWithValue(pk)
		s.Require().NoError(err)
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   math.LegacyOneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec()),
			MinSelfDelegation: math.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), math.LegacyOneDec()))
	}
	s.validators = validators

	// set validators and delegations
	stakingParams := stakingtypes.DefaultParams()
	// set bond demon to be aevmos
	stakingParams.BondDenom = evmosutil.ExampleAttoDenom
	stakingGenesis := stakingtypes.NewGenesisState(stakingParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalBondAmt := bondAmt.Add(bondAmt)
	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens and delegated tokens to total supply
		totalSupply = totalSupply.Add(b.Coins.Add(sdk.NewCoin(evmosutil.ExampleAttoDenom, totalBondAmt))...)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(evmosutil.ExampleAttoDenom, totalBondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{}, []banktypes.SendEnabled{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	s.Require().NoError(err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			ChainId:         evmosutil.ExampleChainID,
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: chainutil.DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)
	app.Commit()

	// instantiate new header
	header := evmosutil.NewHeader(
		2,
		time.Now().UTC(),
		evmosutil.ExampleChainID,
		sdk.ConsAddress(validators[0].GetOperator()),
		tmhash.Sum([]byte("app")),
		tmhash.Sum([]byte("validators")),
	)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// create Context
	s.ctx = app.BaseApp.NewContext(false, header)
	s.app = app
}

func (s *PrecompileTestSuite) DoSetupTest() {
	// generate validator private/public key
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	s.Require().NoError(err)

	privVal2 := mock.NewPV()
	pubKey2, err := privVal2.GetPubKey()
	s.Require().NoError(err)

	// create validator set with two validators
	validator := cmttypes.NewValidator(pubKey, 1)
	validator2 := cmttypes.NewValidator(pubKey2, 2)
	s.valSet = cmttypes.NewValidatorSet([]*cmttypes.Validator{validator, validator2})
	signers := make(map[string]cmttypes.PrivValidator)
	signers[pubKey.Address().String()] = privVal
	signers[pubKey2.Address().String()] = privVal2

	// generate genesis account
	addr, priv := evmosutiltx.NewAddrKey()
	s.privKey = priv
	s.address = addr
	s.signer = evmosutiltx.NewSigner(priv)

	baseAcc := authtypes.NewBaseAccount(priv.PubKey().Address().Bytes(), priv.PubKey(), 0, 0)

	amount := sdk.TokensFromConsensusPower(5, evmostypes.AttoPowerReduction)

	balance := banktypes.Balance{
		Address: baseAcc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(evmosutil.ExampleAttoDenom, amount)),
	}

	s.SetupWithGenesisValSet(s.valSet, []authtypes.GenesisAccount{baseAcc}, balance)

	// Create StateDB
	s.stateDB = statedb.New(s.ctx, s.app.EVMKeeper, statedb.NewEmptyTxConfig(common.BytesToHash(s.ctx.HeaderHash().Bytes())))

	// bond denom
	stakingParams := s.app.StakingKeeper.GetParams(s.ctx)
	stakingParams.BondDenom = evmosutil.ExampleAttoDenom
	s.bondDenom = stakingParams.BondDenom
	err = s.app.StakingKeeper.SetParams(s.ctx, stakingParams)
	s.Require().NoError(err)

	s.ethSigner = ethtypes.LatestSignerForChainID(s.app.EVMKeeper.ChainID())

	precompile, err := distribution.NewPrecompile(s.app.DistrKeeper, *s.app.StakingKeeper, s.app.AuthzKeeper, s.app.EVMKeeper)
	s.Require().NoError(err)
	s.precompile = precompile

	coins := sdk.NewCoins(sdk.NewCoin(evmosutil.ExampleAttoDenom, math.NewInt(5_000_000_000_000_000_000)))
	inflCoins := sdk.NewCoins(sdk.NewCoin(evmosutil.ExampleAttoDenom, math.NewInt(2_000_000_000_000_000_000)))
	distrCoins := sdk.NewCoins(sdk.NewCoin(evmosutil.ExampleAttoDenom, math.NewInt(3_000_000_000_000_000_000)))
	err = s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, coins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, minttypes.ModuleName, authtypes.FeeCollectorName, inflCoins)
	s.Require().NoError(err)
	err = s.app.BankKeeper.SendCoinsFromModuleToModule(s.ctx, minttypes.ModuleName, distrtypes.ModuleName, distrCoins)
	s.Require().NoError(err)

	queryHelperEvm := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	evmtypes.RegisterQueryServer(queryHelperEvm, s.app.EVMKeeper)
	s.queryClientEVM = evmtypes.NewQueryClient(queryHelperEvm)
}

// DeployContract deploys a contract that calls the distribution precompile's methods for testing purposes.
func (s *PrecompileTestSuite) DeployContract(contract evmtypes.CompiledContract) (addr common.Address, err error) {
	addr, err = chainutil.DeployContract(
		s.ctx,
		s.app,
		s.privKey,
		s.queryClientEVM,
		contract,
	)
	return
}

type stakingRewards struct {
	Delegator sdk.AccAddress
	Validator stakingtypes.Validator
	RewardAmt math.Int
}

// prepareStakingRewards prepares the test suite for testing delegation rewards.
//
// Specified rewards amount are allocated to the specified validator using the distribution keeper,
// such that the given amount of tokens is outstanding as a staking reward for the account.
//
// The setup is done in the following way:
//   - Fund the account with the given address with the given rewards amount.
//   - Delegate the rewards amount to the validator specified
//   - Allocate rewards to the validator.
func (s *PrecompileTestSuite) prepareStakingRewards(stkRs ...stakingRewards) {
	for _, r := range stkRs {
		// fund account to make delegation
		err := chainutil.FundAccountWithBaseDenom(s.ctx, s.app.BankKeeper, r.Delegator, r.RewardAmt.Int64())
		s.Require().NoError(err)
		// set distribution module account balance which pays out the rewards
		distrAcc := s.app.DistrKeeper.GetDistributionAccount(s.ctx)
		err = chainutil.FundModuleAccount(s.ctx, s.app.BankKeeper, distrAcc.GetName(), sdk.NewCoins(sdk.NewCoin(s.bondDenom, r.RewardAmt)))
		s.Require().NoError(err)

		// make a delegation
		_, err = s.app.StakingKeeper.Delegate(s.ctx, r.Delegator, r.RewardAmt, stakingtypes.Unspecified, r.Validator, true)
		s.Require().NoError(err)

		// end block to bond validator and increase block height
		sdkstaking.EndBlocker(s.ctx, s.app.StakingKeeper)
		// allocate rewards to validator (of these 50% will be paid out to the delegator)
		allocatedRewards := sdk.NewDecCoins(sdk.NewDecCoin(s.bondDenom, r.RewardAmt.Mul(math.NewInt(2))))
		s.app.DistrKeeper.AllocateTokensToValidator(s.ctx, r.Validator, allocatedRewards)
	}
	s.NextBlock()
}

// NextBlock commits the current block and sets up the next block.
func (s *PrecompileTestSuite) NextBlock() {
	var err error
	s.ctx, err = chainutil.CommitAndCreateNewCtx(s.ctx, s.app, time.Second, s.valSet)
	s.Require().NoError(err)
}

// setupValidatorSlashes sets slashes events for the provided validator
// returns the slash event set
func (s *PrecompileTestSuite) setupValidatorSlashes(valAddr sdk.ValAddress, slashesCount uint64) distrtypes.ValidatorSlashEvent {
	const (
		initialHeight uint64 = 2
		initialPeriod uint64 = 1
	)

	slashEvent := distrtypes.ValidatorSlashEvent{ValidatorPeriod: 1, Fraction: math.LegacyNewDec(5)}

	for i := uint64(0); i < slashesCount; i++ {
		s.app.DistrKeeper.SetValidatorSlashEvent(s.ctx, valAddr, initialHeight+i, initialPeriod+i, slashEvent)
	}

	return slashEvent
}
