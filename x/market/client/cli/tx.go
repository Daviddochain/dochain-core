package cli

import (
	"fmt"
	"strings"

	core "github.com/Daviddochain/dochain-core/v4/types"
	feeutils "github.com/Daviddochain/dochain-core/v4/custom/auth/client/utils"
	"github.com/Daviddochain/dochain-core/v4/x/market/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	marketTxCmd := &cobra.Command{
		Use:                        "market",
		Short:                      "Market transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	marketTxCmd.AddCommand(
		GetSwapCmd(),
	)

	return marketTxCmd
}

// GetSwapCmd creates and broadcasts a MsgSwap or MsgSwapSend transaction.
func GetSwapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [offer-coin] [ask-denom] [to-address]",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Atomically swap assets at the oracle's effective exchange rate",
		Long: strings.TrimSpace(fmt.Sprintf(`
Swap the offer coin into the ask denom at the oracle's effective exchange rate.

Example:
$ dochaind tx market swap "1000%s" "%s"

A recipient address can also be specified. If omitted, the trader is also the recipient.

Example:
$ dochaind tx market swap "1000%s" "%s" "do1..."
`, core.MicroDoDenom, core.MicroDoDenom, core.MicroDoDenom, core.MicroDoDenom)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Generate transaction factory for gas simulation.
			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			offerCoinStr := args[0]
			offerCoin, err := sdk.ParseCoinNormalized(offerCoinStr)
			if err != nil {
				return err
			}

			askDenom := args[1]
			fromAddress := clientCtx.GetFromAddress()

			var msg sdk.Msg
			if len(args) == 3 {
				toAddress, err := sdk.AccAddressFromBech32(args[2])
				if err != nil {
					return err
				}

				msg = types.NewMsgSwapSend(fromAddress, toAddress, offerCoin, askDenom)

				if !clientCtx.GenerateOnly && txf.Fees().IsZero() {
					// Estimate tax and gas.
					stdFee, err := feeutils.ComputeFeesWithCmd(clientCtx, cmd.Flags(), msg)
					if err != nil {
						return err
					}

					// Override gas and fees.
					txf = txf.
						WithFees(stdFee.Amount.String()).
						WithGas(stdFee.Gas).
						WithSimulateAndExecute(false).
						WithGasPrices("")
				}
			} else {
				msg = types.NewMsgSwap(fromAddress, offerCoin, askDenom)
			}

			// Build and sign the transaction, then broadcast it.
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}