// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testdata

import (
	contractutils "github.com/evmos/os/contracts/utils"
	evmtypes "github.com/evmos/os/x/evm/types"
)

func LoadGovCallerContract() (evmtypes.CompiledContract, error) {
	return contractutils.LoadContractFromJSONFile("GovCaller.json")
}
