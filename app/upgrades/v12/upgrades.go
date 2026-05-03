package v12

import (
    "context"

    upgradetypes "cosmossdk.io/x/upgrade/types"
    "github.com/Daviddochain/dochain-core/v4/app/keepers"
    "github.com/Daviddochain/dochain-core/v4/app/upgrades"
    "github.com/cosmos/cosmos-sdk/types/module"
)

func CreateV12UpgradeHandler(
    mm *module.Manager,
    cfg module.Configurator,
    _ upgrades.BaseAppParamManager,
    _ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
    return func(c context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
        return mm.RunMigrations(c, cfg, fromVM)
    }
}
