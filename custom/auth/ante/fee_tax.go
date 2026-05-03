package ante

import sdk "github.com/cosmos/cosmos-sdk/types"

// Tax module removed from DoChain.
// Keep ante fee flow compiling by returning zero taxes.
func FilterMsgAndComputeTax(ctx sdk.Context, msgs ...sdk.Msg) (sdk.Coins, sdk.Coins) {
    return sdk.NewCoins(), sdk.NewCoins()
}
