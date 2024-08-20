package testutil_test

import (
	chainconfig "github.com/evmos/os/example_chain/osd/config"
	"testing"

	"github.com/evmos/os/example_chain"
	"github.com/evmos/os/testutil"
	"github.com/stretchr/testify/require"
)

func TestRequireSameTestDenom(t *testing.T) {
	require.Equal(t,
		testutil.ExampleAttoDenom,
		example_chain.ExampleChainDenom,
		"test denoms should be the same across the repo",
	)
}

func TestRequireSameTestBech32Prefix(t *testing.T) {
	require.Equal(t,
		testutil.ExampleBech32Prefix,
		chainconfig.Bech32Prefix,
		"bech32 prefixes should be the same across the repo",
	)
}

func TestRequireSameWEVMOSMainnet(t *testing.T) {
	require.Equal(t,
		testutil.WEVMOSContractMainnet,
		example_chain.WEVMOSContractMainnet,
		"wevmos contract addresses should be the same across the repo",
	)
}
