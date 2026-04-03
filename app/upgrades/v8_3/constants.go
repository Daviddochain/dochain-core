package v8_3

import (
	"github.com/Daviddochain/do-core/v4/app/upgrades"
)

const UpgradeName = "v8_3"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV83UpgradeHandler,
}





