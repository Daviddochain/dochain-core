package keeper

import (
	"context"
	"math"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/treasury/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// querier is used because Keeper would otherwise have duplicate methods when
// used directly, and gRPC method names take precedence over q.
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the treasury QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

var _ types.QueryServer = querier{}

// Params queries params of the treasury module.
func (q querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

// TaxRate returns the current tax rate.
func (q querier) TaxRate(c context.Context, _ *types.QueryTaxRateRequest) (*types.QueryTaxRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryTaxRateResponse{TaxRate: q.GetTaxRate(ctx)}, nil
}

// TaxCap returns the tax cap of a denom.
func (q querier) TaxCap(c context.Context, req *types.QueryTaxCapRequest) (*types.QueryTaxCapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryTaxCapResponse{TaxCap: q.GetTaxCap(ctx, req.Denom)}, nil
}

// TaxCaps returns all tax caps.
func (q querier) TaxCaps(c context.Context, _ *types.QueryTaxCapsRequest) (*types.QueryTaxCapsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var taxCaps []types.QueryTaxCapsResponseItem
	q.IterateTaxCap(ctx, func(denom string, taxCap sdkmath.Int) bool {
		taxCaps = append(taxCaps, types.QueryTaxCapsResponseItem{
			Denom:  denom,
			TaxCap: taxCap,
		})
		return false
	})

	return &types.QueryTaxCapsResponse{TaxCaps: taxCaps}, nil
}

// RewardWeight returns the current reward weight.
func (q querier) RewardWeight(c context.Context, _ *types.QueryRewardWeightRequest) (*types.QueryRewardWeightResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryRewardWeightResponse{RewardWeight: q.GetRewardWeight(ctx)}, nil
}

// SeigniorageProceeds returns the current seigniorage proceeds.
func (q querier) SeigniorageProceeds(c context.Context, _ *types.QuerySeigniorageProceedsRequest) (*types.QuerySeigniorageProceedsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QuerySeigniorageProceedsResponse{SeigniorageProceeds: q.PeekEpochSeigniorage(ctx)}, nil
}

// TaxProceeds returns the current tax proceeds.
func (q querier) TaxProceeds(c context.Context, _ *types.QueryTaxProceedsRequest) (*types.QueryTaxProceedsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryTaxProceedsResponse{TaxProceeds: q.PeekEpochTaxProceeds(ctx)}, nil
}

// Indicators returns the current TRL information.
func (q querier) Indicators(c context.Context, _ *types.QueryIndicatorsRequest) (*types.QueryIndicatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	// Compute total staked Do (TSL).
	TSL, err := q.stakingKeeper.TotalBondedTokens(sdk.WrapSDKContext(ctx))
	if err != nil {
		TSL = sdkmath.ZeroInt()
	}

	// Compute tax rewards (TR).
	taxRewards := sdk.NewDecCoinsFromCoins(q.PeekEpochTaxProceeds(ctx)...)
	TR := q.alignCoins(ctx, taxRewards, core.MicroDoDenom)

	epoch := q.GetEpoch(ctx)
	var res types.QueryIndicatorsResponse
	if epoch == 0 {
		res = types.QueryIndicatorsResponse{
			TRLYear:  TR.QuoInt(TSL),
			TRLMonth: TR.QuoInt(TSL),
		}
	} else {
		params := q.GetParams(ctx)
		previousEpochCtx := ctx.WithBlockHeight(ctx.BlockHeight() - int64(core.BlocksPerWeek))
		trlYear := q.rollingAverageIndicator(previousEpochCtx, int64(params.WindowLong-1), TRL)
		trlMonth := q.rollingAverageIndicator(previousEpochCtx, int64(params.WindowShort-1), TRL)

		computedEpochForYear := int64(math.Min(float64(params.WindowLong-1), float64(epoch)))
		computedEpochForMonth := int64(math.Min(float64(params.WindowShort-1), float64(epoch)))

		trlYear = trlYear.MulInt64(computedEpochForYear).Add(TR.QuoInt(TSL)).QuoInt64(computedEpochForYear + 1)
		trlMonth = trlMonth.MulInt64(computedEpochForMonth).Add(TR.QuoInt(TSL)).QuoInt64(computedEpochForMonth + 1)

		res = types.QueryIndicatorsResponse{
			TRLYear:  trlYear,
			TRLMonth: trlMonth,
		}
	}

	return &res, nil
}

// BurnTaxExemptionList returns all burn tax exemption addresses.
func (q querier) BurnTaxExemptionList(c context.Context, req *types.QueryBurnTaxExemptionListRequest) (*types.QueryBurnTaxExemptionListResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	sub := prefix.NewStore(ctx.KVStore(q.storeKey), types.BurnTaxExemptionListPrefix)

	var addresses []string
	pageRes, err := query.FilteredPaginate(sub, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		address := string(key)
		addresses = append(addresses, address)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBurnTaxExemptionListResponse{
		Addresses:  addresses,
		Pagination: pageRes,
	}, nil
}