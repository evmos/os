package keeper_test

import (
	"reflect"

	exampleapp "github.com/evmos/os/example_chain"
	"github.com/evmos/os/testutil"
	"github.com/evmos/os/x/evm/types"
)

func (suite *KeeperTestSuite) TestParams() {
	params := types.DefaultParamsWithEVMDenom(testutil.ExampleAttoDenom)

	testCases := []struct {
		name      string
		paramsFun func() interface{}
		getFun    func() interface{}
		expected  bool
	}{
		{
			"success - Checks if the default params are set correctly",
			func() interface{} {
				// NOTE: we are using the EVM genesis state for the example app here, because
				// we have different assumptions for the evmOS offering and the example chain.
				return exampleapp.NewEVMGenesisState().Params
			},
			func() interface{} {
				return suite.app.EVMKeeper.GetParams(suite.ctx)
			},
			true,
		},
		{
			"success - EvmDenom param is set to \"inj\" and can be retrieved correctly",
			func() interface{} {
				params.EvmDenom = "inj"
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.EvmDenom
			},
			func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetEvmDenom()
			},
			true,
		},
		{
			"success - Check Access Control Create param is set to restricted and can be retrieved correctly",
			func() interface{} {
				params.AccessControl = types.AccessControl{
					Create: types.AccessControlType{
						AccessType: types.AccessTypeRestricted,
					},
				}
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return types.AccessTypeRestricted
			},
			func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetAccessControl().Create.AccessType
			},
			true,
		},
		{
			"success - Check Access control param is set to restricted and can be retrieved correctly",
			func() interface{} {
				params.AccessControl = types.AccessControl{
					Call: types.AccessControlType{
						AccessType: types.AccessTypeRestricted,
					},
				}
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return types.AccessTypeRestricted
			},
			func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetAccessControl().Call.AccessType
			},
			true,
		},
		{
			"success - Check AllowUnprotectedTxs param is set to false and can be retrieved correctly",
			func() interface{} {
				params.AllowUnprotectedTxs = false
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.AllowUnprotectedTxs
			},
			func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetAllowUnprotectedTxs()
			},
			true,
		},
		{
			"success - Check ChainConfig param is set to the default value and can be retrieved correctly",
			func() interface{} {
				params.ChainConfig = types.DefaultChainConfig()
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err)
				return params.ChainConfig
			},
			func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetChainConfig()
			},
			true,
		},
		{
			name: "success - Active precompiles are sorted when setting params",
			paramsFun: func() interface{} {
				params.ActiveStaticPrecompiles = []string{
					"0x0000000000000000000000000000000000000801",
					"0x0000000000000000000000000000000000000800",
				}
				err := suite.app.EVMKeeper.SetParams(suite.ctx, params)
				suite.Require().NoError(err, "expected no error when setting params")

				// NOTE: return sorted slice here because the precompiles should be sorted when setting the params
				return []string{
					"0x0000000000000000000000000000000000000800",
					"0x0000000000000000000000000000000000000801",
				}
			},
			getFun: func() interface{} {
				evmParams := suite.app.EVMKeeper.GetParams(suite.ctx)
				return evmParams.GetActiveStaticPrecompiles()
			},
			expected: true,
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			outcome := reflect.DeepEqual(tc.paramsFun(), tc.getFun())
			suite.Require().Equal(tc.expected, outcome)
		})
	}
}
