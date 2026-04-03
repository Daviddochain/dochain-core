package v7

import (
	store "cosmossdk.io/store/types"
	"github.com/Daviddochain/do-core/v4/app/upgrades"
	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10/types"
)

const UpgradeName = "v7"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV7UpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ibchookstypes.StoreKey,
		},
	},
}





