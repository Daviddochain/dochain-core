package types

import (
	"encoding/json"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates a new GenesisState object.
func NewGenesisState(doPoolDelta math.LegacyDec, params Params) *GenesisState {
	return &GenesisState{
		DoPoolDelta: doPoolDelta,
		Params:      params,
	}
}

// DefaultGenesisState returns the default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		DoPoolDelta: math.LegacyZeroDec(),
		Params:      DefaultParams(),
	}
}

// ValidateGenesis validates the provided market genesis state.
func ValidateGenesis(data *GenesisState) error {
	return data.Params.Validate()
}

// GetGenesisStateFromAppState returns x/market GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}