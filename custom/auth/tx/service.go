package tx

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    grpc1 "github.com/cosmos/gogoproto/grpc"
)

type TreasuryService struct{}

func NewTreasuryService() TreasuryService {
    return TreasuryService{}
}

func (ts TreasuryService) ComputeTax(
    goCtx context.Context,
    req *ComputeTaxRequest,
) (*ComputeTaxResponse, error) {
    _ = sdk.UnwrapSDKContext(goCtx)
    _ = req

    return &ComputeTaxResponse{
        TaxAmount: sdk.NewCoins(),
    }, nil
}

func RegisterTxService(
    server grpc1.Server,
) {
    RegisterServiceServer(server, NewTreasuryService())
}
