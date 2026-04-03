package market

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	core "github.com/Daviddochain/do-core/v4/types"
	"github.com/Daviddochain/do-core/v4/x/market/keeper"
	"github.com/Daviddochain/do-core/v4/x/market/types"
)

var randomPrice = sdkmath.LegacyNewDec(1700)

func setup(t *testing.T) (keeper.TestInput, types.MsgServer) {
	input := keeper.CreateTestInput(t)

	params := input.MarketKeeper.GetParams(input.Ctx)
	input.MarketKeeper.SetParams(input.Ctx, params)
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroSDRDenom, randomPrice)
	input.OracleKeeper.SetLunaExchangeRate(input.Ctx, core.MicroKRWDenom, randomPrice)
	h := keeper.NewMsgServerImpl(input.MarketKeeper)

	return input, h
}





