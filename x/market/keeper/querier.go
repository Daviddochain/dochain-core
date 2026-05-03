package keeper

import (
	"context"

	"github.com/Daviddochain/dochain-core/v4/x/market/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// querier is used because Keeper would otherwise have duplicate methods when
// used directly, and gRPC method names take precedence over q.
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the market QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

var _ types.QueryServer = querier{}

// Params queries the parameters of the market module.
func (q querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

// Swap simulates a market swap query.
func (q querier) Swap(c context.Context, req *types.QuerySwapRequest) (*types.QuerySwapResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if err := sdk.ValidateDenom(req.AskDenom); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid ask denom")
	}

	offerCoin, err := sdk.ParseCoinNormalized(req.OfferCoin)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	retCoin, err := q.simulateSwap(ctx, offerCoin, req.AskDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QuerySwapResponse{ReturnCoin: retCoin}, nil
}

// DoPoolDelta queries the current Do pool delta.
func (q querier) DoPoolDelta(c context.Context, _ *types.QueryDoPoolDeltaRequest) (*types.QueryDoPoolDeltaResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	doPoolDelta := q.GetDoPoolDelta(ctx)
	return &types.QueryDoPoolDeltaResponse{DoPoolDelta: doPoolDelta}, nil
}