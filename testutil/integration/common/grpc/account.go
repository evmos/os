// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)
package grpc

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/os/encoding"
	exampleapp "github.com/evmos/os/example_chain"
)

// GetAccount returns the account for the given address.
func (gqh *IntegrationHandler) GetAccount(address string) (authtypes.AccountI, error) {
	authClient := gqh.network.GetAuthClient()
	res, err := authClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nil, err
	}

	encodingCgf := encoding.MakeConfig(exampleapp.ModuleBasics)
	var acc authtypes.AccountI
	if err = encodingCgf.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, err
	}
	return acc, nil
}
