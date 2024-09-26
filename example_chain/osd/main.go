// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	examplechain "github.com/evmos/os/example_chain"
	"github.com/evmos/os/example_chain/osd/cmd"
	chainconfig "github.com/evmos/os/example_chain/osd/config"
)

func main() {
	setupSDKConfig()
	chainconfig.RegisterDenoms()

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "osd", examplechain.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}

func setupSDKConfig() {
	config := sdk.GetConfig()
	chainconfig.SetBech32Prefixes(config)
	config.Seal()
}
