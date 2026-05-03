package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/stretchr/testify/require"
)

func TestDoPoolDeltaUpdate(t *testing.T) {
	input := CreateTestInput(t)

	doPoolDelta := input.MarketKeeper.GetDoPoolDelta(input.Ctx)
	require.Equal(t, sdkmath.LegacyZeroDec(), doPoolDelta)

	diff := sdkmath.LegacyNewDec(10)
	input.MarketKeeper.SetDoPoolDelta(input.Ctx, diff)

	doPoolDelta = input.MarketKeeper.GetDoPoolDelta(input.Ctx)
	require.Equal(t, diff, doPoolDelta)
}

// TestReplenishPools tests that
// each pools move towards base pool
func TestReplenishPools(t *testing.T) {
	input := CreateTestInput(t)
	input.OracleKeeper.SetDoExchangeRate(input.Ctx, core.MicroSDRDenom, sdkmath.LegacyOneDec())

	basePool := input.MarketKeeper.BasePool(input.Ctx)
	doPoolDelta := input.MarketKeeper.GetDoPoolDelta(input.Ctx)
	require.True(t, doPoolDelta.IsZero())

	// Positive delta
	diff := basePool.QuoInt64((int64)(core.BlocksPerDay))
	input.MarketKeeper.SetDoPoolDelta(input.Ctx, diff)

	input.MarketKeeper.ReplenishPools(input.Ctx)

	doPoolDelta = input.MarketKeeper.GetDoPoolDelta(input.Ctx)
	replenishAmt := diff.QuoInt64((int64)(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx)))
	expectedDelta := diff.Sub(replenishAmt)
	require.Equal(t, expectedDelta, doPoolDelta)

	// Negative delta
	diff = diff.Neg()
	input.MarketKeeper.SetDoPoolDelta(input.Ctx, diff)

	input.MarketKeeper.ReplenishPools(input.Ctx)

	doPoolDelta = input.MarketKeeper.GetDoPoolDelta(input.Ctx)
	replenishAmt = diff.QuoInt64((int64)(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx)))
	expectedDelta = diff.Sub(replenishAmt)
	require.Equal(t, expectedDelta, doPoolDelta)
}






