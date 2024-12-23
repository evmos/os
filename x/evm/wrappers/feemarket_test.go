package wrappers_test

import (
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

func TestGetBaseFee(t *testing.T) {
	testCases := []struct {
		name        string
		evmDenom    string
		evmDecimals uint8
		expResult   *big.Int
		mockSetup   func(*testutil.MockFeeMarketKeeper)
	}{
		{
			name:        "success - does not convert 18 decimals",
			evmDenom:    TokenDenom,
			evmDecimals: 18,
			expResult:   big.NewInt(1e18), // 1 token in 18 decimals
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyNewDec(1e18))
			},
		},
		{
			name:        "success - convert 6 decimals to 18 decimals",
			evmDenom:    TokenDenom,
			evmDecimals: 6,
			expResult:   big.NewInt(1e18), // 1 token in 18 decimals
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyNewDec(1_000_000))
			},
		},
		{
			name:        "success - nil base fee",
			evmDenom:    TokenDenom,
			evmDecimals: 6,
			expResult:   nil,
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyDec{})
			},
		},
		{
			name:        "success - small amount 18 decimals",
			evmDenom:    TokenDenom,
			evmDecimals: 6,
			expResult:   big.NewInt(1e12), // 0.000001 token in 18 decimals
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyNewDec(1))
			},
		},
		{
			name:        "success - base fee is zero",
			evmDenom:    TokenDenom,
			evmDecimals: 6,
			expResult:   big.NewInt(0),
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyNewDec(0))
			},
		},
		{
			name:        "success - truncate decimals with number less than 1",
			evmDenom:    TokenDenom,
			evmDecimals: 6,
			expResult:   big.NewInt(0), // 0.000001 token in 18 decimals
			mockSetup: func(mfk *testutil.MockFeeMarketKeeper) {
				mfk.EXPECT().
					GetBaseFee(gomock.Any()).
					Return(sdkmath.LegacyNewDecWithPrec(1, 13)) // multiplied by 1e12 is still less than 1
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

			ctrl := gomock.NewController(t)
			mockFeeMarketKeeper := testutil.NewMockFeeMarketKeeper(ctrl)
			tc.mockSetup(mockFeeMarketKeeper)

			feeMarketWrapper := wrappers.NewFeeMarketWrapper(mockFeeMarketKeeper)
			result := feeMarketWrapper.GetBaseFee(sdk.Context{})

			require.Equal(t, tc.expResult, result)
		})
	}
}
