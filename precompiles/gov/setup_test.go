package gov_test

import (
	"testing"
	"time"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/evmos/os/precompiles/gov"
	testconstants "github.com/evmos/os/testutil/constants"
	"github.com/evmos/os/testutil/integration/os/factory"
	"github.com/evmos/os/testutil/integration/os/grpc"
	testkeyring "github.com/evmos/os/testutil/integration/os/keyring"
	"github.com/evmos/os/testutil/integration/os/network"

	"github.com/stretchr/testify/suite"
)

type PrecompileTestSuite struct {
	suite.Suite

	network     *network.UnitTestNetwork
	factory     factory.TxFactory
	grpcHandler grpc.Handler
	keyring     testkeyring.Keyring

	precompile *gov.Precompile
}

func TestPrecompileUnitTestSuite(t *testing.T) {
	suite.Run(t, new(PrecompileTestSuite))
}

func (s *PrecompileTestSuite) SetupTest() {
	keyring := testkeyring.New(2)

	// seed the db with one proposal
	customGen := network.CustomGenesisState{}
	now := time.Now().UTC()
	inOneHour := now.Add(time.Hour)

	var err error
	anyMessage, err := types.NewAnyWithValue(TestProposalMsgs[0])
	if err != nil {
		panic(err)
	}
	prop := &govv1.Proposal{
		Id:              1,
		Status:          govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:      &now,
		DepositEndTime:  &inOneHour,
		VotingStartTime: &now,
		FinalTallyResult: &govv1.TallyResult{
			YesCount:        "0",
			AbstainCount:    "0",
			NoCount:         "0",
			NoWithVetoCount: "0",
		},
		VotingEndTime: &inOneHour,
		Metadata:      "ipfs://CID",
		Title:         "test prop",
		Summary:       "test prop",
		Proposer:      keyring.GetAccAddr(0).String(),
		Messages:      []*types.Any{anyMessage},
	}

	prop2 := &govv1.Proposal{
		Id:              2,
		Status:          govv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD,
		SubmitTime:      &now,
		DepositEndTime:  &inOneHour,
		VotingStartTime: &now,
		FinalTallyResult: &govv1.TallyResult{
			YesCount:        "0",
			AbstainCount:    "0",
			NoCount:         "0",
			NoWithVetoCount: "0",
		},
		VotingEndTime: &inOneHour,
		Metadata:      "ipfs://CID",
		Title:         "test prop",
		Summary:       "test prop",
		Proposer:      keyring.GetAccAddr(1).String(),
		Messages:      []*types.Any{anyMessage},
	}

	bankGen := banktypes.DefaultGenesisState()
	bankGen.Balances = []banktypes.Balance{{
		Address: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(testconstants.ExampleAttoDenom, math.NewInt(200))),
	}}
	govGen := govv1.DefaultGenesisState()
	govGen.Deposits = []*govv1.Deposit{
		{
			ProposalId: 1,
			Depositor:  keyring.GetAccAddr(0).String(),
			Amount:     sdk.NewCoins(sdk.NewCoin(testconstants.ExampleAttoDenom, math.NewInt(100))),
		},
		{
			ProposalId: 2,
			Depositor:  keyring.GetAccAddr(1).String(),
			Amount:     sdk.NewCoins(sdk.NewCoin(testconstants.ExampleAttoDenom, math.NewInt(100))),
		},
	}
	govGen.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin(testconstants.ExampleAttoDenom, math.NewInt(100)))
	govGen.Proposals = append(govGen.Proposals, prop)
	govGen.Proposals = append(govGen.Proposals, prop2)
	customGen[govtypes.ModuleName] = govGen
	customGen[banktypes.ModuleName] = bankGen

	nw := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
		network.WithCustomGenesis(customGen),
	)
	grpcHandler := grpc.NewIntegrationHandler(nw)
	txFactory := factory.New(nw, grpcHandler)

	s.factory = txFactory
	s.grpcHandler = grpcHandler
	s.keyring = keyring
	s.network = nw

	if s.precompile, err = gov.NewPrecompile(
		s.network.App.GovKeeper,
		s.network.App.AuthzKeeper,
	); err != nil {
		panic(err)
	}
}
