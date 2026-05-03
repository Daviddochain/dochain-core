package simulation

import (
    "cosmossdk.io/math"
    appparams "github.com/Daviddochain/dochain-core/v4/app/params"
    "github.com/cosmos/cosmos-sdk/types/module"
    authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// RandomizedGenState generates a random GenesisState for auth
func RandomizedGenState(simState *module.SimulationState, _ authtypes.RandomGenesisAccountsFn) {
    _ = math.LegacyZeroDec()
    _ = appparams.BondDenom
}
