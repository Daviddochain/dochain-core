package market

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/Daviddochain/dochain-core/v4/x/market/keeper"
	"github.com/stretchr/testify/require"
)

func TestReplenishPools(t *testing.T) {
	input := keeper.CreateTestInput(t)

	doDelta := sdkmath.LegacyNewDecWithPrec(17987573223725367, 3)
	input.MarketKeeper.SetDoPoolDelta(input.Ctx, doDelta)

	for i := 0; i < 100; i++ {
		doDelta = input.MarketKeeper.GetDoPoolDelta(input.Ctx)

		poolRecoveryPeriod := int64(input.MarketKeeper.PoolRecoveryPeriod(input.Ctx))
		doRegressionAmt := doDelta.QuoInt64(poolRecoveryPeriod)

		EndBlocker(input.Ctx, input.MarketKeeper)

		doPoolDelta := input.MarketKeeper.GetDoPoolDelta(input.Ctx)
		require.Equal(t, doDelta.Sub(doRegressionAmt), doPoolDelta)
	}
}






