// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package example_chain

import (
	"fmt"
	"maps"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/common"
	bankprecompile "github.com/evmos/os/precompiles/bank"
	"github.com/evmos/os/precompiles/bech32"
	distprecompile "github.com/evmos/os/precompiles/distribution"
	govprecompile "github.com/evmos/os/precompiles/gov"
	ics20precompile "github.com/evmos/os/precompiles/ics20"
	"github.com/evmos/os/precompiles/p256"
	stakingprecompile "github.com/evmos/os/precompiles/staking"
	erc20Keeper "github.com/evmos/os/x/erc20/keeper"
	"github.com/evmos/os/x/evm/core/vm"
	evmkeeper "github.com/evmos/os/x/evm/keeper"
	transferkeeper "github.com/evmos/os/x/ibc/transfer/keeper"
)

const bech32PrecompileBaseGas = 6_000

// NewAvailableStaticPrecompiles returns the list of all available static precompiled contracts from evmOS.
//
// NOTE: this should only be used during initialization of the Keeper.
func NewAvailableStaticPrecompiles(
	stakingKeeper stakingkeeper.Keeper,
	distributionKeeper distributionkeeper.Keeper,
	bankKeeper bankkeeper.Keeper,
	erc20Keeper erc20Keeper.Keeper,
	authzKeeper authzkeeper.Keeper,
	transferKeeper transferkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	evmKeeper *evmkeeper.Keeper,
	govKeeper govkeeper.Keeper,
) map[common.Address]vm.PrecompiledContract {
	// Clone the mapping from the latest EVM fork.
	precompiles := maps.Clone(vm.PrecompiledContractsBerlin)

	// secp256r1 precompile as per EIP-7212
	p256Precompile := &p256.Precompile{}

	bech32Precompile, err := bech32.NewPrecompile(bech32PrecompileBaseGas)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bech32 precompile: %w", err))
	}

	stakingPrecompile, err := stakingprecompile.NewPrecompile(stakingKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate staking precompile: %w", err))
	}

	distributionPrecompile, err := distprecompile.NewPrecompile(
		distributionKeeper,
		stakingKeeper,
		authzKeeper,
		evmKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate distribution precompile: %w", err))
	}

	ibcTransferPrecompile, err := ics20precompile.NewPrecompile(
		stakingKeeper,
		transferKeeper,
		channelKeeper,
		authzKeeper,
		evmKeeper,
	)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate ICS20 precompile: %w", err))
	}

	bankPrecompile, err := bankprecompile.NewPrecompile(bankKeeper, erc20Keeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate bank precompile: %w", err))
	}

	govPrecompile, err := govprecompile.NewPrecompile(govKeeper, authzKeeper)
	if err != nil {
		panic(fmt.Errorf("failed to instantiate gov precompile: %w", err))
	}

	// Stateless precompiles
	precompiles[bech32Precompile.Address()] = bech32Precompile
	precompiles[p256Precompile.Address()] = p256Precompile

	// Stateful precompiles
	precompiles[stakingPrecompile.Address()] = stakingPrecompile
	precompiles[distributionPrecompile.Address()] = distributionPrecompile
	precompiles[ibcTransferPrecompile.Address()] = ibcTransferPrecompile
	precompiles[bankPrecompile.Address()] = bankPrecompile
	precompiles[govPrecompile.Address()] = govPrecompile

	return precompiles
}
