package keeper

import (
	sdkmath "cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/treasury/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SettleSeigniorage computes seigniorage and distributes it between burn and the distribution
// module community pool.
func (k Keeper) SettleSeigniorage(ctx sdk.Context) {
	// Mint seigniorage for burn and community pool distribution.
	seigniorageDoAmt := k.PeekEpochSeigniorage(ctx)
	if seigniorageDoAmt.LTE(sdkmath.ZeroInt()) {
		return
	}

	// Settle current epoch seigniorage.
	rewardWeight := k.GetRewardWeight(ctx)

	// Represent seigniorage in the chain's base Do denom.
	seigniorageDecCoin := sdk.NewDecCoin(core.MicroDoDenom, seigniorageDoAmt)

	// Mint seigniorage.
	seigniorageCoin, _ := seigniorageDecCoin.TruncateDecimal()
	seigniorageCoins := sdk.NewCoins(seigniorageCoin)
	if seigniorageCoins.IsValid() {
		if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, seigniorageCoins); err != nil {
			panic(err)
		}
	}
	seigniorageAmt := seigniorageCoin.Amount

	// Burn the configured reward portion.
	burnAmt := rewardWeight.MulInt(seigniorageAmt).TruncateInt()
	burnCoins := sdk.NewCoins(sdk.NewCoin(core.MicroDoDenom, burnAmt))
	if burnCoins.IsValid() {
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
			panic(err)
		}
	}

	// Send the remainder to the distribution module.
	leftAmt := seigniorageAmt.Sub(burnAmt)
	leftCoins := sdk.NewCoins(sdk.NewCoin(core.MicroDoDenom, leftAmt))
	if leftCoins.IsValid() {
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			k.distributionModuleName,
			leftCoins,
		); err != nil {
			panic(err)
		}

		// Update distribution community pool.
		feePool, err := k.distrKeeper.FeePool.Get(ctx)
		if err != nil {
			panic(err)
		}
		feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(leftCoins...)...)
		k.distrKeeper.FeePool.Set(ctx, feePool)
	}
}