package keeper

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestSettle(t *testing.T) {
	input := CreateTestInput(t)

	faucetBalance := input.BankKeeper.GetBalance(input.Ctx, input.AccountKeeper.GetModuleAddress(faucetAccountName), core.MicroDoDenom)
	burnAmt := sdkmath.NewInt(rand.Int63()%faucetBalance.Amount.Int64() + 1)
	initialDoSupply := input.BankKeeper.GetSupply(input.Ctx, core.MicroDoDenom)
	input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek))
	err := input.BankKeeper.BurnCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroDoDenom, burnAmt)))
	require.NoError(t, err)

	// check seigniorage update
	require.Equal(t, burnAmt, input.TreasuryKeeper.PeekEpochSeigniorage(input.Ctx))

	input.TreasuryKeeper.SettleSeigniorage(input.Ctx)
	doSupply := input.BankKeeper.GetSupply(input.Ctx, core.MicroDoDenom)
	feePool, _ := input.DistrKeeper.FeePool.Get(input.Ctx)

	// Reward weight portion of seigniorage burned
	rewardWeight := input.TreasuryKeeper.GetRewardWeight(input.Ctx)
	communityPoolAmt := burnAmt.Sub(rewardWeight.MulInt(burnAmt).TruncateInt())

	require.Equal(t, doSupply.Amount, initialDoSupply.Amount.Sub(burnAmt).Add(communityPoolAmt))
	require.Equal(t, communityPoolAmt, feePool.CommunityPool.AmountOf(core.MicroDoDenom).TruncateInt())
}

func TestOneRewardWeightSettle(t *testing.T) {
	input := CreateTestInput(t)

	// set zero reward weight
	input.TreasuryKeeper.SetRewardWeight(input.Ctx, sdkmath.LegacyOneDec())

	faucetBalance := input.BankKeeper.GetBalance(input.Ctx, input.AccountKeeper.GetModuleAddress(faucetAccountName), core.MicroDoDenom)
	burnAmt := sdkmath.NewInt(rand.Int63()%faucetBalance.Amount.Int64() + 1)
	initialDoSupply := input.BankKeeper.GetSupply(input.Ctx, core.MicroDoDenom)
	input.TreasuryKeeper.RecordEpochInitialIssuance(input.Ctx)

	input.Ctx = input.Ctx.WithBlockHeight(int64(core.BlocksPerWeek))
	err := input.BankKeeper.BurnCoins(input.Ctx, faucetAccountName, sdk.NewCoins(sdk.NewCoin(core.MicroDoDenom, burnAmt)))
	require.NoError(t, err)

	// check seigniorage update
	require.Equal(t, burnAmt, input.TreasuryKeeper.PeekEpochSeigniorage(input.Ctx))

	input.TreasuryKeeper.SettleSeigniorage(input.Ctx)
	doSupply := input.BankKeeper.GetSupply(input.Ctx, core.MicroDoDenom)
	feePool, _ := input.DistrKeeper.FeePool.Get(input.Ctx)

	// Reward weight portion of seigniorage burned
	require.Equal(t, doSupply.Amount, initialDoSupply.Amount.Sub(burnAmt))
	require.Equal(t, sdkmath.ZeroInt(), feePool.CommunityPool.AmountOf(core.MicroDoDenom).TruncateInt())
}






