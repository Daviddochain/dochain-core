package wasmbinding

import (
	//	"github.com/Daviddochain/dochain-core/v4/wasmbinding/bindings"
	marketkeeper "github.com/Daviddochain/dochain-core/v4/x/market/keeper"
	oraclekeeper "github.com/Daviddochain/dochain-core/v4/x/oracle/keeper"
	treasurykeeper "github.com/Daviddochain/dochain-core/v4/x/treasury/keeper"
)

type QueryPlugin struct {
	marketKeeper   *marketkeeper.Keeper
	oracleKeeper   *oraclekeeper.Keeper
	treasuryKeeper *treasurykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(tmk *marketkeeper.Keeper, tok *oraclekeeper.Keeper, ttk *treasurykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		marketKeeper:   tmk,
		oracleKeeper:   tok,
		treasuryKeeper: ttk,
	}
}






