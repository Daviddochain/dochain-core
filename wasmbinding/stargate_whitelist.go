package wasmbinding

import (
	"fmt"
	"sync"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v3/types"
	markettypes "github.com/Daviddochain/dochain-core/v4/x/market/types"
	oracletypes "github.com/Daviddochain/dochain-core/v4/x/oracle/types"
	treasurytypes "github.com/Daviddochain/dochain-core/v4/x/treasury/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

// stargateWhitelist keeps whitelist and its deterministic
// response binding for stargate queries.
//
// The query can be multi-thread, so we have to use
// thread safe sync.Map.
var stargateWhitelist sync.Map

func init() {
	// market
	setWhitelistedQuery("/do.market.v1beta1.Query/Swap", &markettypes.QuerySwapResponse{})

	// treasury
	setWhitelistedQuery("/do.treasury.v1beta1.Query/TaxCap", &treasurytypes.QueryTaxCapResponse{})
	setWhitelistedQuery("/do.treasury.v1beta1.Query/TaxRate", &treasurytypes.QueryTaxRateResponse{})

	// oracle
	setWhitelistedQuery("/do.oracle.v1beta1.Query/ExchangeRate", &oracletypes.QueryExchangeRateResponse{})
}

// GetWhitelistedQuery returns the whitelisted query at the provided path.
// If the query does not exist, or it was setup wrong by the chain, this returns an error.
func GetWhitelistedQuery(queryPath string) (codec.ProtoMarshaler, error) {
	protoResponseAny, isWhitelisted := stargateWhitelist.Load(queryPath)
	if !isWhitelisted {
		return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("'%s' path is not allowed from the contract", queryPath)}
	}
	protoResponseType, ok := protoResponseAny.(codec.ProtoMarshaler)
	if !ok {
		return nil, wasmvmtypes.Unknown{}
	}
	return protoResponseType, nil
}

func setWhitelistedQuery(queryPath string, protoType codec.ProtoMarshaler) {
	stargateWhitelist.Store(queryPath, protoType)
}






