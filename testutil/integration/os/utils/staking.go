// Copyright Tharsis Labs Ltd.(Evmos)
// SPDX-License-Identifier:ENCL-1.0(https://github.com/evmos/evmos/blob/main/LICENSE)

package utils

import (
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/evmos/os/testutil/integration/os/grpc"
	"github.com/evmos/os/testutil/integration/os/network"
)

// WaitToAccrueRewards is a helper function that waits for rewards to
// accumulate up to a specified expected amount
func WaitToAccrueRewards(n network.Network, gh grpc.Handler, delegatorAddr string, expRewards sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err     error
		lapse   = time.Hour * 24 * 7 // one week
		rewards = sdk.DecCoins{}
	)

	if err = checkNonZeroInflation(n); err != nil {
		return nil, err
	}

	expAmt := expRewards.AmountOf(n.GetBaseDenom())
	for rewards.AmountOf(n.GetBaseDenom()).LT(expAmt) {
		rewards, err = checkRewardsAfter(n, gh, delegatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking rewards")
		}
	}

	return rewards, err
}

// checkRewardsAfter is a helper function that checks the accrued rewards
// after the provided timelapse
func checkRewardsAfter(n network.Network, gh grpc.Handler, delegatorAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetDelegationTotalRewards(delegatorAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Total, nil
}

// WaitToAccrueCommission is a helper function that waits for commission to
// accumulate up to a specified expected amount
func WaitToAccrueCommission(n network.Network, gh grpc.Handler, validatorAddr string, expCommission sdk.DecCoins) (sdk.DecCoins, error) {
	var (
		err        error
		lapse      = time.Hour * 24 * 7 // one week
		commission = sdk.DecCoins{}
	)

	if err := checkNonZeroInflation(n); err != nil {
		return nil, err
	}

	expAmt := expCommission.AmountOf(n.GetBaseDenom())
	for commission.AmountOf(n.GetBaseDenom()).LT(expAmt) {
		commission, err = checkCommissionAfter(n, gh, validatorAddr, lapse)
		if err != nil {
			return nil, errorsmod.Wrap(err, "error checking commission")
		}
	}

	return commission, err
}

// checkCommissionAfter is a helper function that checks the accrued commission
// after the provided time lapse
func checkCommissionAfter(n network.Network, gh grpc.Handler, valAddr string, lapse time.Duration) (sdk.DecCoins, error) {
	err := n.NextBlockAfter(lapse)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to commit block after voting period ends")
	}

	res, err := gh.GetValidatorCommission(valAddr)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error while querying for delegation rewards")
	}

	return res.Commission.Commission, nil
}

// checkNonZeroInflation is a helper function that checks if the network's
// inflation is non-zero.
// This is required to ensure that rewards and commission are accrued.
func checkNonZeroInflation(n network.Network) error {
	res, err := n.GetMintClient().Inflation(n.GetContext(), &minttypes.QueryInflationRequest{})
	if err != nil {
		return errorsmod.Wrap(err, "failed to get inflation")
	}

	if res.Inflation.IsZero() {
		return errors.New("inflation is zero; must be non-zero for rewards or commission to be distributed")
	}

	return nil
}
