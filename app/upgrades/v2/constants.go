package v2

import (
	store "cosmossdk.io/store/types"
	"github.com/Daviddochain/do-core/v4/app/upgrades"
)

const UpgradeName = "v2"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV2UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{},
}





