package v10_1

import (
    store "cosmossdk.io/store/types"
    "github.com/Daviddochain/dochain-core/v4/app/upgrades"
)

const UpgradeName = "v10_1"

var Upgrade = upgrades.Upgrade{
    UpgradeName:          UpgradeName,
    CreateUpgradeHandler: CreateV101UpgradeHandler,
    StoreUpgrades: store.StoreUpgrades{
        Added: []string{},
    },
}
