// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	erc20types "github.com/evmos/os/x/erc20/types"
)

var ExampleTokenPairs = []erc20types.TokenPair{
	{
		Erc20Address:  WEVMOSContractMainnet,
		Denom:         ExampleAttoDenom,
		Enabled:       true,
		ContractOwner: erc20types.OWNER_MODULE,
	},
}
