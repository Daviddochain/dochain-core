package simulation

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/Daviddochain/dochain-core/v4/x/market/keeper"
	"github.com/Daviddochain/dochain-core/v4/x/market/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/stretchr/testify/require"
)

func TestDecodeDistributionStore(t *testing.T) {
	cdc := keeper.MakeTestCodec(t)
	dec := NewDecodeStore(cdc)

	doDelta := sdkmath.LegacyNewDecWithPrec(12, 2)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.DoPoolDeltaKey, Value: cdc.MustMarshal(&sdk.DecProto{Dec: doDelta})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"DoPoolDelta", fmt.Sprintf("%v\n%v", doDelta, doDelta)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}






