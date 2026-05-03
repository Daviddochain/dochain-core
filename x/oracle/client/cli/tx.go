package cli

import (
	"fmt"
	"strings"

	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	oracleTxCmd := &cobra.Command{
		Use:                        "oracle",
		Short:                      "Oracle transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	oracleTxCmd.AddCommand(
		GetCmdDelegateFeederPermission(),
		GetCmdAggregateDoRatePrevote(),
		GetCmdAggregateDoRateVote(),
	)

	return oracleTxCmd
}

// GetCmdDelegateFeederPermission creates a feeder permission delegation tx and signs it with the given key.
func GetCmdDelegateFeederPermission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-feeder [feeder]",
		Args:  cobra.ExactArgs(1),
		Short: "Delegate the permission to vote for the oracle to an address",
		Long: strings.TrimSpace(`
Delegate the permission to submit exchange rate votes for the oracle to an address.

Delegation can keep your validator operator key offline and use a separate replaceable key online.

$ dochaind tx oracle set-feeder do1...

where "do1..." is the address you want to delegate your voting rights to.
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			voter := clientCtx.GetFromAddress()
			validator := sdk.ValAddress(voter)

			feederStr := args[0]
			feeder, err := sdk.AccAddressFromBech32(feederStr)
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgDelegateFeedConsent(validator, feeder)}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAggregateDoRatePrevote creates an aggregate exchange rate prevote tx and signs it with the given key.
func GetCmdAggregateDoRatePrevote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate-prevote [salt] [exchange-rates] [validator]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Submit an oracle aggregate prevote for the exchange rates of Do",
		Long: strings.TrimSpace(fmt.Sprintf(`
Submit an oracle aggregate prevote for the exchange rates of Do denominated in one or more denoms.
The purpose of the aggregate prevote is to hide the aggregate exchange rate vote with a hash formatted
as the SHA256 hex string of "{salt}:{exchange_rate}{denom},...,{exchange_rate}{denom}:{voter}".

Example:
$ dochaind tx oracle aggregate-prevote 1234 1.000000000000000000%[1]s

where "%[1]s" is an example exchange rate tuple using the chain base denom.

If voting from a voting delegate, set "validator" to the address of the validator to vote on behalf of:
$ dochaind tx oracle aggregate-prevote 1234 1.000000000000000000%[1]s dovaloper1...
`, core.MicroDoDenom)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			salt := args[0]
			exchangeRatesStr := args[1]
			_, err = types.ParseDoRateTuples(exchangeRatesStr)
			if err != nil {
				return fmt.Errorf("given exchange_rates {%s} is not a valid format; exchange_rate should be formatted as DecCoins; %s", exchangeRatesStr, err.Error())
			}

			voter := clientCtx.GetFromAddress()
			validator := sdk.ValAddress(voter)

			if len(args) == 3 {
				parsedVal, err := sdk.ValAddressFromBech32(args[2])
				if err != nil {
					return errors.Wrap(err, "validator address is invalid")
				}
				validator = parsedVal
			}

			hash := types.GetAggregateVoteHash(salt, exchangeRatesStr, validator)
			msgs := []sdk.Msg{types.NewMsgAggregateDoRatePrevote(hash, voter, validator)}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetCmdAggregateDoRateVote creates an aggregate exchange rate vote tx and signs it with the given key.
func GetCmdAggregateDoRateVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregate-vote [salt] [exchange-rates] [validator]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Submit an oracle aggregate vote for the exchange rates of Do",
		Long: strings.TrimSpace(fmt.Sprintf(`
Submit an aggregate vote for the exchange rates of Do with respect to the input denom.
This is the companion to a prevote submitted in the previous vote period.

Example:
$ dochaind tx oracle aggregate-vote 1234 1.000000000000000000%[1]s

where "%[1]s" is an example exchange rate tuple using the chain base denom.

"salt" should match the salt used to generate the SHA256 hex in the aggregate prevote.

If voting from a voting delegate, set "validator" to the address of the validator to vote on behalf of:
$ dochaind tx oracle aggregate-vote 1234 1.000000000000000000%[1]s dovaloper1...
`, core.MicroDoDenom)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			salt := args[0]
			exchangeRatesStr := args[1]
			_, err = types.ParseDoRateTuples(exchangeRatesStr)
			if err != nil {
				return fmt.Errorf("given exchange_rate {%s} is not a valid format; exchange rate should be formatted as DecCoin; %s", exchangeRatesStr, err.Error())
			}

			voter := clientCtx.GetFromAddress()
			validator := sdk.ValAddress(voter)

			if len(args) == 3 {
				parsedVal, err := sdk.ValAddressFromBech32(args[2])
				if err != nil {
					return errors.Wrap(err, "validator address is invalid")
				}
				validator = parsedVal
			}

			msgs := []sdk.Msg{types.NewMsgAggregateDoRateVote(salt, exchangeRatesStr, voter, validator)}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgs...)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}