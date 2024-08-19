// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package example_chain

import (
	"encoding/json"

	"cosmossdk.io/simapp"
	"github.com/evmos/os/encoding"
	evmtypes "github.com/evmos/os/x/evm/types"
)

// GenesisState of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState() simapp.GenesisState {
	encCfg := encoding.MakeConfig(ModuleBasics)

	genesisState := ModuleBasics.DefaultGenesis(encCfg.Codec)

	// define new chain-specific EVM genesis state with correct EVM denom
	evmGenesis := evmtypes.DefaultGenesisState()
	evmGenesis.Params.EvmDenom = ExampleChainDenom
	genesisState[evmtypes.ModuleName] = encCfg.Codec.MustMarshalJSON(evmGenesis)

	return genesisState
}
