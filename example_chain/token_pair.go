// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package example_chain

import erc20types "github.com/evmos/os/x/erc20/types"

// WEVMOSContractMainnet is the WEVMOS contract address for mainnet
const WEVMOSContractMainnet = "0xD4949664cD82660AaE99bEdc034a0deA8A0bd517"

// ExampleTokenPairs creates a slice of token pairs, that contains a pair for the native denom of the example chain
// implementation.
var ExampleTokenPairs = []erc20types.TokenPair{
	{
		Erc20Address:  WEVMOSContractMainnet,
		Denom:         ExampleChainDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE,
	},
}
