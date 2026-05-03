package keeper

import (
	sdkmath "cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpoch returns the current epoch from the current block height.
func (k Keeper) GetEpoch(ctx sdk.Context) int64 {
	return ctx.BlockHeight() / int64(core.BlocksPerWeek)
}

//
// Computes important economic indicators for the stability of Do.
//
// Three important concepts:
// - MR: fees + seigniorage for a given epoch sum to mining rewards
// - SR: computes the seigniorage reward
// - TR: computes the tax reward
// - TSL: total staked Do
// - TRL: computes the tax reward per unit Do (TR / TSL)

// alignCoins aligns the provided coins to the given denom through the market swap.
func (k Keeper) alignCoins(ctx sdk.Context, coins sdk.DecCoins, denom string) (alignedAmt sdkmath.LegacyDec) {
	alignedAmt = sdkmath.LegacyZeroDec()

	for _, coinReward := range coins {
		if coinReward.Denom != denom {
			swappedReward, err := k.marketKeeper.ComputeInternalSwap(ctx, coinReward, denom)
			if err != nil {
				continue
			}
			alignedAmt = alignedAmt.Add(swappedReward.Amount)
		} else {
			alignedAmt = alignedAmt.Add(coinReward.Amount)
		}
	}

	return alignedAmt
}

// UpdateIndicators updates internal indicators.
func (k Keeper) UpdateIndicators(ctx sdk.Context) {
	epoch := k.GetEpoch(ctx)

	// Compute total staked Do (TSL).
	totalStakedDo, err := k.stakingKeeper.TotalBondedTokens(sdk.WrapSDKContext(ctx))
	if err != nil {
		totalStakedDo = sdkmath.ZeroInt()
	}

	k.SetTSL(ctx, epoch, totalStakedDo)

	// Compute tax rewards (TR).
	taxRewards := sdk.NewDecCoinsFromCoins(k.PeekEpochTaxProceeds(ctx)...)
	TR := k.alignCoins(ctx, taxRewards, core.MicroDoDenom)
	k.SetTR(ctx, epoch, TR)

	// Reset tax proceeds after computing TR for the next epoch.
	k.SetEpochTaxProceeds(ctx, sdk.Coins{})

	// Compute seigniorage rewards (SR).
	seigniorage := k.PeekEpochSeigniorage(ctx)
	seigniorageRewardsAmt := k.GetRewardWeight(ctx).MulInt(seigniorage)
	seigniorageRewards := sdk.DecCoins{sdk.NewDecCoinFromDec(core.MicroDoDenom, seigniorageRewardsAmt)}
	SR := k.alignCoins(ctx, seigniorageRewards, core.MicroDoDenom)

	k.SetSR(ctx, epoch, SR)
}

// TRL returns tax rewards per Do for the epoch.
func TRL(ctx sdk.Context, epoch int64, k Keeper) sdkmath.LegacyDec {
	tr := k.GetTR(ctx, epoch)
	tsl := k.GetTSL(ctx, epoch)

	// division by zero protection
	if tr.IsZero() || tsl.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	return tr.QuoInt(tsl)
}

// SR returns seigniorage rewards for the epoch.
func SR(ctx sdk.Context, epoch int64, k Keeper) sdkmath.LegacyDec {
	return k.GetSR(ctx, epoch)
}

// MR returns mining rewards = seigniorage rewards + tax rewards for the epoch.
func MR(ctx sdk.Context, epoch int64, k Keeper) sdkmath.LegacyDec {
	return k.GetTR(ctx, epoch).Add(k.GetSR(ctx, epoch))
}

// sumIndicator returns the sum of the indicator over several epochs.
// If current epoch < epochs, it returns the best available partial sum.
func (k Keeper) sumIndicator(
	ctx sdk.Context,
	epochs int64,
	indicator func(ctx sdk.Context, epoch int64, k Keeper) sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	sum := sdkmath.LegacyZeroDec()
	curEpoch := k.GetEpoch(ctx)

	for i := curEpoch; i >= 0 && i > (curEpoch-epochs); i-- {
		val := indicator(ctx, i, k)
		sum = sum.Add(val)
	}

	return sum
}

// rollingAverageIndicator returns the rolling average of the indicator over several epochs.
// If current epoch < epochs, it returns the best available partial average.
func (k Keeper) rollingAverageIndicator(
	ctx sdk.Context,
	epochs int64,
	indicator func(ctx sdk.Context, epoch int64, k Keeper) sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	sum := sdkmath.LegacyZeroDec()
	curEpoch := k.GetEpoch(ctx)

	var i int64
	for i = curEpoch; i >= 0 && i > (curEpoch-epochs); i-- {
		val := indicator(ctx, i, k)
		sum = sum.Add(val)
	}

	computedEpochs := curEpoch - i
	if computedEpochs == 0 {
		return sum
	}

	return sum.QuoInt64(computedEpochs)
}