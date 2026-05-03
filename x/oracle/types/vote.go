package types

import (
    "fmt"
    "strings"

    "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "gopkg.in/yaml.v2"
)

// NewAggregateDoRatePrevote returns AggregateDoRatePrevote object
func NewAggregateDoRatePrevote(hash AggregateVoteHash, voter sdk.ValAddress, submitBlock uint64) AggregateDoRatePrevote {
    return AggregateDoRatePrevote{
        Hash:        hash.String(),
        Voter:       voter.String(),
        SubmitBlock: submitBlock,
    }
}

// String implement stringify
func (v AggregateDoRatePrevote) String() string {
    out, _ := yaml.Marshal(v)
    return string(out)
}

// NewAggregateDoRateVote creates a AggregateDoRateVote instance
func NewAggregateDoRateVote(exchangeRateTuples DoRateTuples, voter sdk.ValAddress) AggregateDoRateVote {
    return AggregateDoRateVote{
        ExchangeRateTuples: exchangeRateTuples,
        Voter:              voter.String(),
    }
}

// String implement stringify
func (v AggregateDoRateVote) String() string {
    out, _ := yaml.Marshal(v)
    return string(out)
}

// NewDoRateTuple creates a DoRateTuple instance
func NewDoRateTuple(denom string, exchangeRate math.LegacyDec) DoRateTuple {
    return DoRateTuple{
        Denom:        denom,
        ExchangeRate: exchangeRate,
    }
}

// String implement stringify
func (v DoRateTuple) String() string {
    out, _ := yaml.Marshal(v)
    return string(out)
}

// DoRateTuples - array of DoRateTuple
type DoRateTuples []DoRateTuple

// String implements fmt.Stringer interface
func (tuples DoRateTuples) String() string {
    out, _ := yaml.Marshal(tuples)
    return string(out)
}

// ParseDoRateTuples DoRateTuple parser
func ParseDoRateTuples(tuplesStr string) (DoRateTuples, error) {
    tuplesStr = strings.TrimSpace(tuplesStr)
    if len(tuplesStr) == 0 {
        return nil, nil
    }

    tupleStrs := strings.Split(tuplesStr, ",")
    tuples := make(DoRateTuples, len(tupleStrs))
    duplicateCheckMap := make(map[string]bool)

    for i, tupleStr := range tupleStrs {
        decCoin, err := sdk.ParseDecCoin(tupleStr)
        if err != nil {
            return nil, err
        }

        tuples[i] = DoRateTuple{
            Denom:        decCoin.Denom,
            ExchangeRate: decCoin.Amount,
        }

        if _, ok := duplicateCheckMap[decCoin.Denom]; ok {
            return nil, fmt.Errorf("duplicated denom %s", decCoin.Denom)
        }
        duplicateCheckMap[decCoin.Denom] = true
    }

    return tuples, nil
}
