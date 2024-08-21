package keeper_test

import (
	"reflect"

	"github.com/evmos/os/testutil"
	"github.com/evmos/os/x/erc20/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := suite.app.Erc20Keeper.GetParams(suite.ctx)
	suite.app.Erc20Keeper.SetParams(suite.ctx, params) //nolint:errcheck

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				erc20Params := types.DefaultParams()
				// NOTE: we need to add the example token pair address which is not in the default params but in the genesis state
				// of the test suite app and therefore is returned by the query client.
				erc20Params.NativePrecompiles = append(erc20Params.NativePrecompiles, testutil.WEVMOSContractMainnet)

				return erc20Params
			},
			func() interface{} {
				return suite.app.Erc20Keeper.GetParams(suite.ctx)
			},
			true,
		},
		{
			"success - Checks if dynamic precompiles are set correctly",
			func() interface{} {
				params.DynamicPrecompiles = []string{"0xB5124FA2b2cF92B2D469b249433BA1c96BDF536D", "0xC4CcDf91b810a61cCB48b35ccCc066C63bf94B4F"}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.DynamicPrecompiles
			},
			func() interface{} {
				return suite.app.Erc20Keeper.GetParams(suite.ctx).DynamicPrecompiles
			},
			true,
		},
		{
			"success - Checks if native precompiles are set correctly",
			func() interface{} {
				params.NativePrecompiles = []string{"0x205CF44075E77A3543abC690437F3b2819bc450a", "0x8FA78CEB7F04118Ec6d06AaC37Ca854691d8e963"}
				err := suite.app.Erc20Keeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.NativePrecompiles
			},
			func() interface{} {
				return suite.app.Erc20Keeper.GetParams(suite.ctx).NativePrecompiles
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}
