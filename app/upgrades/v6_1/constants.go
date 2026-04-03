package v61

import (
	store "cosmossdk.io/store/types"
	"github.com/Daviddochain/do-core/v4/app/upgrades"
)

const UpgradeName = "v6_1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV6_1UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}





