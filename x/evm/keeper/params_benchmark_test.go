package keeper_test

import (
	"testing"

	"github.com/evmos/os/x/evm/types"
)

func BenchmarkSetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTestWithT(b)
	params := types.DefaultParams()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.app.EVMKeeper.SetParams(suite.ctx, params)
	}
}

func BenchmarkGetParams(b *testing.B) {
	suite := KeeperTestSuite{}
	suite.SetupTestWithT(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.app.EVMKeeper.GetParams(suite.ctx)
	}
}
