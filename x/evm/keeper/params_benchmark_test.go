package keeper_test

import (
	"testing"

	testconstants "github.com/evmos/os/testutil/constants"
	"github.com/evmos/os/x/evm/types"
)

func BenchmarkSetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTest()
	params := types.DefaultParamsWithEVMDenom(testconstants.ExampleAttoDenom)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.network.App.EVMKeeper.SetParams(suite.network.GetContext(), params)
	}
}

func BenchmarkGetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTest()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.network.App.EVMKeeper.GetParams(suite.network.GetContext())
	}
}
