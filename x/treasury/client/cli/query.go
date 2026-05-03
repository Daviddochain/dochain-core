package cli

import (
	"context"
	"fmt"
	"strings"

	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/treasury/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the CLI query commands for this module.
func GetQueryCmd() *cobra.Command {
	treasuryQueryCmd := &cobra.Command{
		Use:                        "treasury",
		Short:                      "Querying commands for the treasury module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	treasuryQueryCmd.AddCommand(
		GetCmdQueryTaxRate(),
		GetCmdQueryTaxCap(),
		GetCmdQueryTaxCaps(),
		GetCmdQueryRewardWeight(),
		GetCmdQueryTaxProceeds(),
		GetCmdQuerySeigniorageProceeds(),
		GetCmdQueryIndicators(),
		GetCmdQueryParams(),
		GetCmdQueryExemptlist(),
	)

	return treasuryQueryCmd
}

// GetCmdQueryTaxRate implements the tax-rate query command.
func GetCmdQueryTaxRate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-rate",
		Args:  cobra.NoArgs,
		Short: "Query the current tax rate",
		Long: strings.TrimSpace(`
Query the current tax rate of the epoch.

$ dochaind query treasury tax-rate
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TaxRate(context.Background(), &types.QueryTaxRateRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryTaxCap implements the tax-cap query command.
func GetCmdQueryTaxCap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-cap [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query the current tax cap for a denom",
		Long: strings.TrimSpace(fmt.Sprintf(`
Query the current tax cap for a denom.
The tax levied on a transaction is at most the configured tax cap, regardless of
the transaction size.

Example:
$ dochaind query treasury tax-cap %s
`, core.MicroDoDenom)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			denom := args[0]

			res, err := queryClient.TaxCap(context.Background(), &types.QueryTaxCapRequest{
				Denom: denom,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryTaxCaps implements the tax-caps query command.
func GetCmdQueryTaxCaps() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-caps",
		Args:  cobra.NoArgs,
		Short: "Query the current tax caps for all denoms",
		Long: strings.TrimSpace(`
Query the current tax caps for all denoms.
The tax levied on a transaction is at most the configured tax cap, regardless of
the transaction size.

$ dochaind query treasury tax-caps
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TaxCaps(context.Background(), &types.QueryTaxCapsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryRewardWeight implements the reward-weight query command.
func GetCmdQueryRewardWeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reward-weight",
		Args:  cobra.NoArgs,
		Short: "Query the current reward weight",
		Long: strings.TrimSpace(`
Query the reward weight of the current epoch.

$ dochaind query treasury reward-weight
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.RewardWeight(context.Background(), &types.QueryRewardWeightRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryTaxProceeds implements the tax-proceeds query command.
func GetCmdQueryTaxProceeds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-proceeds",
		Args:  cobra.NoArgs,
		Short: "Query the tax proceeds for the current epoch",
		Long: strings.TrimSpace(`
Query the tax proceeds corresponding to the current epoch.
The return value is an sdk.Coins set of all taxes collected.

$ dochaind query treasury tax-proceeds
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.TaxProceeds(context.Background(), &types.QueryTaxProceedsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQuerySeigniorageProceeds implements the seigniorage-proceeds query command.
func GetCmdQuerySeigniorageProceeds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seigniorage-proceeds",
		Args:  cobra.NoArgs,
		Short: "Query the seigniorage proceeds for the current epoch",
		Long: strings.TrimSpace(fmt.Sprintf(`
Query the seigniorage proceeds corresponding to the current epoch.
The return value will be in units of '%s'.

$ dochaind query treasury seigniorage-proceeds
`, core.MicroDoDenom)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.SeigniorageProceeds(context.Background(), &types.QuerySeigniorageProceedsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryIndicators implements the indicators query command.
func GetCmdQueryIndicators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indicators",
		Args:  cobra.NoArgs,
		Short: "Query the current treasury indicators",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Indicators(context.Background(), &types.QueryIndicatorsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current treasury parameters",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdQueryExemptlist queries all burn tax exemption addresses.
func GetCmdQueryExemptlist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-tax-exemption-list",
		Args:  cobra.NoArgs,
		Short: "Query all burn tax exemption addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.BurnTaxExemptionList(
				context.Background(),
				&types.QueryBurnTaxExemptionListRequest{Pagination: pageReq},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "burn tax exemption list")
	return cmd
}