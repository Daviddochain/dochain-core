//nolint:revive
package v13_1

import (
	"github.com/Daviddochain/do-core/v4/app/upgrades"
)

const UpgradeName = "v13_1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV131UpgradeHandler,
}





