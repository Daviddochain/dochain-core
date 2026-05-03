package v12

import (
    store "cosmossdk.io/store/types"
    "github.com/Daviddochain/dochain-core/v4/app/upgrades"
)

const UpgradeName = "v12"

var Upgrade = upgrades.Upgrade{
    UpgradeName:          UpgradeName,
    CreateUpgradeHandler: CreateV12UpgradeHandler,
    StoreUpgrades: store.StoreUpgrades{
        Added: []string{},
    },
}
