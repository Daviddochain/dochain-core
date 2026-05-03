package types_test

import (
	"testing"

	"github.com/Daviddochain/dochain-core/v4/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestParseDoRateTuples(t *testing.T) {
	valid := "123.0udo,123.123ukrw"
	_, err := types.ParseDoRateTuples(valid)
	require.NoError(t, err)

	duplicatedDenom := "100.0udo,123.123ukrw,121233.123ukrw"
	_, err = types.ParseDoRateTuples(duplicatedDenom)
	require.Error(t, err)

	invalidCoins := "123.123"
	_, err = types.ParseDoRateTuples(invalidCoins)
	require.Error(t, err)

	invalidCoinsWithValid := "123.0udo,123.1"
	_, err = types.ParseDoRateTuples(invalidCoinsWithValid)
	require.Error(t, err)

	abstainCoinsWithValid := "0.0udo,123.1ukrw"
	_, err = types.ParseDoRateTuples(abstainCoinsWithValid)
	require.NoError(t, err)
}






