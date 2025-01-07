// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package evm_test

import (
	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evmante "github.com/evmos/os/ante/evm"
	commonfactory "github.com/evmos/os/testutil/integration/common/factory"
	testfactory "github.com/evmos/os/testutil/integration/os/factory"
	"github.com/evmos/os/testutil/integration/os/grpc"
	testkeyring "github.com/evmos/os/testutil/integration/os/keyring"
	"github.com/evmos/os/testutil/integration/os/network"
)

func (suite *EvmAnteTestSuite) TestUpdateCumulativeGasWanted() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)

	testCases := []struct {
		name                string
		msgGasWanted        uint64
		maxTxGasWanted      uint64
		cumulativeGasWanted uint64
		getCtx              func() sdktypes.Context
		expectedResponse    uint64
	}{
		{
			name:                "when is NOT checkTx and cumulativeGasWanted is 0, returns msgGasWanted",
			msgGasWanted:        100,
			maxTxGasWanted:      150,
			cumulativeGasWanted: 0,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(false)
			},
			expectedResponse: 100,
		},
		{
			name:                "when is NOT checkTx and cumulativeGasWanted has value, returns cumulativeGasWanted + msgGasWanted",
			msgGasWanted:        100,
			maxTxGasWanted:      150,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(false)
			},
			expectedResponse: 150,
		},
		{
			name:                "when is checkTx, maxTxGasWanted is not 0 and msgGasWanted > maxTxGasWanted, returns cumulativeGasWanted + maxTxGasWanted",
			msgGasWanted:        200,
			maxTxGasWanted:      100,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(true)
			},
			expectedResponse: 150,
		},
		{
			name:                "when is checkTx, maxTxGasWanted is not 0 and msgGasWanted < maxTxGasWanted, returns cumulativeGasWanted + msgGasWanted",
			msgGasWanted:        50,
			maxTxGasWanted:      100,
			cumulativeGasWanted: 50,
			getCtx: func() sdktypes.Context {
				return unitNetwork.GetContext().WithIsCheckTx(true)
			},
			expectedResponse: 100,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function under test
			gasWanted := evmante.UpdateCumulativeGasWanted(
				tc.getCtx(),
				tc.msgGasWanted,
				tc.maxTxGasWanted,
				tc.cumulativeGasWanted,
			)

			suite.Require().Equal(tc.expectedResponse, gasWanted)
		})
	}
}

// NOTE: claim rewards are not tested since there is an independent suite to test just that
func (suite *EvmAnteTestSuite) TestConsumeGasAndEmitEvent() {
	keyring := testkeyring.New(1)
	unitNetwork := network.NewUnitTestNetwork(
		network.WithPreFundedAccounts(keyring.GetAllAccAddrs()...),
	)
	grpcHandler := grpc.NewIntegrationHandler(unitNetwork)
	factory := testfactory.New(unitNetwork, grpcHandler)

	testCases := []struct {
		name          string
		expectedError string
		fees          sdktypes.Coins
		getSender     func() sdktypes.AccAddress
	}{
		{
			name: "success: fees are zero and event emitted",
			fees: sdktypes.Coins{},
			getSender: func() sdktypes.AccAddress {
				// Return prefunded sender
				return keyring.GetKey(0).AccAddr
			},
		},
		{
			name: "success: there are non zero fees, user has sufficient bank balances and event emitted",
			fees: sdktypes.Coins{
				sdktypes.NewCoin(unitNetwork.GetBaseDenom(), sdkmath.NewInt(1000)),
			},
			getSender: func() sdktypes.AccAddress {
				// Return prefunded sender
				return keyring.GetKey(0).AccAddr
			},
		},
		{
			name:          "fail: insufficient user balance, event is NOT emitted",
			expectedError: "failed to deduct transaction costs from user balance",
			fees: sdktypes.Coins{
				sdktypes.NewCoin(unitNetwork.GetBaseDenom(), sdkmath.NewInt(1000)),
			},
			getSender: func() sdktypes.AccAddress {
				// Set up account with too little balance (but not zero)
				index := keyring.AddKey()
				acc := keyring.GetKey(index)

				sender := keyring.GetKey(0)
				_, err := factory.ExecuteCosmosTx(sender.Priv, commonfactory.CosmosTxArgs{
					Msgs: []sdktypes.Msg{&banktypes.MsgSend{
						FromAddress: sender.AccAddr.String(),
						ToAddress:   acc.AccAddr.String(),
						Amount:      sdktypes.Coins{sdktypes.NewCoin(unitNetwork.GetBaseDenom(), sdkmath.NewInt(500))},
					}},
				})
				suite.Require().NoError(err, "failed to send funds to new key")

				return acc.AccAddr
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			sender := tc.getSender()
			prevBalance, err := grpcHandler.GetAllBalances(
				sender,
			)
			suite.Require().NoError(err)

			// Function under test
			err = evmante.ConsumeFeesAndEmitEvent(
				unitNetwork.GetContext(),
				unitNetwork.App.EVMKeeper,
				tc.fees,
				sender,
			)

			if tc.expectedError != "" {
				suite.Require().Error(err)
				suite.Contains(err.Error(), tc.expectedError)

				// Check events are not present
				events := unitNetwork.GetContext().EventManager().Events()
				suite.Require().Zero(len(events))
			} else {
				suite.Require().NoError(err)

				// Check fees are deducted
				afterBalance, err := grpcHandler.GetAllBalances(
					sender,
				)
				suite.Require().NoError(err)
				expectedBalance := prevBalance.Balances.Sub(tc.fees...)
				suite.Require().True(
					expectedBalance.Equal(afterBalance.Balances),
				)

				// Event to be emitted
				expectedEvent := sdktypes.NewEvent(
					sdktypes.EventTypeTx,
					sdktypes.NewAttribute(sdktypes.AttributeKeyFee, tc.fees.String()),
				)
				// Check events are present
				events := unitNetwork.GetContext().EventManager().Events()
				suite.Require().NotZero(len(events))
				suite.Require().Contains(
					events,
					expectedEvent,
				)
			}

			// Reset the context
			err = unitNetwork.NextBlock()
			suite.Require().NoError(err)
		})
	}
}
