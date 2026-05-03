package helper

import (
	oracleexported "github.com/Daviddochain/dochain-core/v4/x/oracle/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func IsOracleTx(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		switch msg.(type) {
		case *oracleexported.MsgAggregateDoRatePrevote:
			continue
		case *oracleexported.MsgAggregateDoRateVote:
			continue
		default:
			return false
		}
	}

	return true
}






