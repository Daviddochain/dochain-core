package v11_1

import (
	"github.com/Daviddochain/do-core/v4/app/upgrades"
)

const UpgradeName = "v11_1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV111UpgradeHandler,
}





