package wrappers_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evmtypes "github.com/evmos/os/x/evm/types"
	"github.com/evmos/os/x/evm/wrappers"
	"github.com/evmos/os/x/evm/wrappers/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestMintAmountToAccount(t *testing.T) {
	tokenDenom := "token"

	testCases := []struct {
		name        string
		evmDenom    string
		evmDecimals uint8
		amount      *big.Int
		recipient   sdk.AccAddress
		expectErr   string
		mockSetup   func(*testutil.MockBankKeeper)
	}{
		{
			name:        "success - convert 18 decimals amount to 6 decimals",
			evmDenom:    tokenDenom,
			evmDecimals: 6,
			amount:      big.NewInt(1e18), // 1 token in 18 decimals
			recipient:   sdk.AccAddress([]byte("test_address")),
			expectErr:   "",
			mockSetup: func(mbk *testutil.MockBankKeeper) {
				expectedCoin := sdk.NewCoin(tokenDenom, sdkmath.NewInt(1e6)) // 1 token in 6 decimals
				expectedCoins := sdk.NewCoins(expectedCoin)

				mbk.EXPECT().
					MintCoins(gomock.Any(), evmtypes.ModuleName, expectedCoins).
					Return(nil)

				mbk.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						evmtypes.ModuleName,
						sdk.AccAddress([]byte("test_address")),
						expectedCoins,
					).Return(nil)
			},
		},
		{
			name:        "success - 18 decimals amount not modified",
			evmDenom:    tokenDenom,
			evmDecimals: 18,
			amount:      big.NewInt(1e18), // 1 token in 18 decimals
			recipient:   sdk.AccAddress([]byte("test_address")),
			expectErr:   "",
			mockSetup: func(mbk *testutil.MockBankKeeper) {
				expectedCoin := sdk.NewCoin(tokenDenom, sdkmath.NewInt(1e18))
				expectedCoins := sdk.NewCoins(expectedCoin)

				mbk.EXPECT().
					MintCoins(gomock.Any(), evmtypes.ModuleName, expectedCoins).
					Return(nil)

				mbk.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						evmtypes.ModuleName,
						sdk.AccAddress([]byte("test_address")),
						expectedCoins,
					).Return(nil)
			},
		},
		{
			name:        "fail - mint coins error",
			evmDenom:    tokenDenom,
			evmDecimals: 6,
			amount:      big.NewInt(1e18),
			recipient:   sdk.AccAddress([]byte("test_address")),
			expectErr:   "failed to mint coins to account in bank wrapper",
			mockSetup: func(mbk *testutil.MockBankKeeper) {
				expectedCoin := sdk.NewCoin(tokenDenom, sdkmath.NewInt(1e6))
				expectedCoins := sdk.NewCoins(expectedCoin)

				mbk.EXPECT().
					MintCoins(gomock.Any(), evmtypes.ModuleName, expectedCoins).
					Return(errors.New("mint error"))
			},
		},
		{
			name:        "fail - send coins error",
			evmDenom:    "evm",
			evmDecimals: 6,
			amount:      big.NewInt(1e18),
			recipient:   sdk.AccAddress([]byte("test_address")),
			expectErr:   "send error",
			mockSetup: func(mbk *testutil.MockBankKeeper) {
				expectedCoin := sdk.NewCoin("evm", sdkmath.NewInt(1e6))
				expectedCoins := sdk.NewCoins(expectedCoin)

				mbk.EXPECT().
					MintCoins(gomock.Any(), evmtypes.ModuleName, expectedCoins).
					Return(nil)

				mbk.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						evmtypes.ModuleName,
						sdk.AccAddress([]byte("test_address")),
						expectedCoins,
					).Return(errors.New("send error"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup EVM configurator to have access to the EVM coin info.
			configurator := evmtypes.NewEVMConfigurator()
			configurator.ResetTestConfig()
			err := configurator.WithEVMCoinInfo(tc.evmDenom, tc.evmDecimals).Configure()
			require.NoError(t, err, "failed to configure EVMConfigurator")

			// Setup mock controller
			ctrl := gomock.NewController(t)

			mockBankKeeper := testutil.NewMockBankKeeper(ctrl)
			tc.mockSetup(mockBankKeeper)

			bankWrapper := wrappers.NewBankWrapper(mockBankKeeper)
			err = bankWrapper.MintAmountToAccount(context.Background(), tc.recipient, tc.amount)

			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
