// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package testutil

import (
	"math/big"

	"cosmossdk.io/math"
)

const (
	// ExampleDenom provides an example denom for use in tests
	ExampleAttoDenom = "aevmos"

	// ExampleBech32Prefix provides an example Bech32 prefix for use in tests
	ExampleBech32Prefix = "evmos"

	// ExampleChainID provides a chain ID that can be used in tests
	ExampleChainID = "evmos_9000-1"

	// DefaultGasPrice is used in testing as the default to use for transactions
	DefaultGasPrice = 20
)

var (
	// AttoPowerReduction defines the power reduction for att units (1e18)
	AttoPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	// MicroPowerReduction defines the power reduction for micro units (1e6)
	MicroPowerReduction = math.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil))
)
