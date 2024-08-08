package main

import (
	"os"

	"github.com/cosmos/cosmos-sdk/server"
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
		switch e := err.(type) {
		case server.ErrorCode:
			os.Exit(e.Code)

		default:
			os.Exit(1)
		}
	}
}

func setupSDKConfig() {
	config := sdk.GetConfig()
	chainconfig.SetBech32Prefixes(config)
	config.Seal()
}
