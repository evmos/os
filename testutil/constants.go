// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

const (
	// ExampleDenom provides an example denom for use in tests
	ExampleAttoDenom = "aevmos"

	// ExampleBech32Prefix provides an example Bech32 prefix for use in tests
	ExampleBech32Prefix = "evmos"

	// ExampleChainIDPrefix provides a chain ID prefix for EIP-155 that can be used in tests
	ExampleChainIDPrefix = "evmos_9000"

	// ExampleChainID provides a chain ID that can be used in tests
	ExampleChainID = ExampleChainIDPrefix + "-1"

	// DefaultGasPrice is used in testing as the default to use for transactions
	DefaultGasPrice = 20

	// WEVMOSContractMainnet is the WEVMOS contract address for mainnet
	WEVMOSContractMainnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"
	// WEVMOSContractTestnet is the WEVMOS contract address for testnet
	WEVMOSContractTestnet = "0xcc491f589b45d4a3c679016195b3fb87d7848210"
)
