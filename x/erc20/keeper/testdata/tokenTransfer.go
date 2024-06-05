// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	_ "embed" // embed compiled smart contract

	contractutils "github.com/evmos/evmos/v18/contracts/utils"
	evmtypes "github.com/evmos/evmos/v18/x/evm/types"
)

// LoadTokenTransferContract loads the tokenTransfer contract.
func LoadTokenTransferContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("tokenTransfer.json")
}
