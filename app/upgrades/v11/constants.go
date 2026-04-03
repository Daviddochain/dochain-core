package v11

import (
	"github.com/Daviddochain/dochain-core/v4/app/upgrades"
)

const UpgradeName = "v11"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV11UpgradeHandler,
}



