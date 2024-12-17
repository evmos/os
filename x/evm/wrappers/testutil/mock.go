package testutil

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/os/x/evm/types"
)

var _ types.BankKeeper = &MockBank{}

type MockBank struct {
	Balances      map[string]sdk.Coin
	HasPermission bool
}

func NewMockBank() *MockBank {
	return &MockBank{
		HasPermission: true,
	}
}

// SendCoinsFromModuleToAccount implements types.BankKeeper.
func (m *MockBank) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	evmCoin := amt[0]

	senderBalance := m.Balances[senderModule]
	if senderBalance.Amount.LT(evmCoin.Amount) {
		return fmt.Errorf("insufficient balance: %s < %s", senderBalance.Amount, evmCoin.Amount)
	}

	m.Balances[recipientAddr.String()] = m.Balances[recipientAddr.String()].Add(evmCoin)
	m.Balances[senderModule] = m.Balances[senderModule].Sub(evmCoin)

	return nil
}

// SendCoinsFromAccountToModule implements types.BankKeeper.
func (m *MockBank) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	evmCoin := amt[0]

	senderBalance := m.Balances[senderAddr.String()]
	if senderBalance.Amount.LT(evmCoin.Amount) {
		return fmt.Errorf("insufficient balance: %s < %s", senderBalance.Amount, evmCoin.Amount)
	}

	m.Balances[recipientModule] = m.Balances[recipientModule].Add(evmCoin)
	m.Balances[senderAddr.String()] = m.Balances[senderAddr.String()].Sub(evmCoin)

	return nil
}

// BurnCoins implements types.BankKeeper.
func (m *MockBank) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	evmCoin := amt[0]

	moduleBalance := m.Balances[moduleName]
	if moduleBalance.Amount.LT(evmCoin.Amount) {
		return fmt.Errorf("insufficient balance: %s < %s", moduleBalance.Amount, evmCoin.Amount)
	}

	m.Balances[moduleName] = m.Balances[moduleName].Sub(evmCoin)

	return nil
}

// MintCoins implements types.BankKeeper.
func (m *MockBank) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	if !m.HasPermission {
		return fmt.Errorf("permission denied")
	}

	evmCoin := amt[0]

	m.Balances[moduleName] = m.Balances[moduleName].Add(evmCoin)

	return nil
}

// GetBalance implements types.BankKeeper.
func (m *MockBank) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return m.Balances[addr.String()]
}

// NOTE: Below methods are not implemented because are not used from the wrapper but are required to
// implement the interface.

// GetAllBalances implements types.BankKeeper.
func (m *MockBank) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	panic("unimplemented")
}

// IsSendEnabledCoins implements types.BankKeeper.
func (m *MockBank) IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error {
	panic("unimplemented")
}

// SendCoins implements types.BankKeeper.
func (m *MockBank) SendCoins(ctx context.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) error {
	panic("unimplemented")
}
