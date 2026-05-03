package utils

import (
    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/client/tx"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
    "github.com/spf13/pflag"

    "cosmossdk.io/math"
)

type (
    EstimateFeeResp struct {
        Fee legacytx.StdFee `json:"fee" yaml:"fee"`
    }
)

type ComputeReqParams struct {
    Memo          string
    ChainID       string
    AccountNumber uint64
    Sequence      uint64
    GasPrices     sdk.DecCoins
    Gas           string
    GasAdjustment string
    Msgs          []sdk.Msg
}

func ComputeFeesWithCmd(
    clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg,
) (*legacytx.StdFee, error) {
    txf, err := tx.NewFactoryCLI(clientCtx, flagSet)
    if err != nil {
        return nil, err
    }

    gas := txf.Gas()
    if txf.SimulateAndExecute() {
        txf, err := prepareFactory(clientCtx, txf)
        if err != nil {
            return nil, err
        }

        _, adj, err := tx.CalculateGas(clientCtx, txf, msgs...)
        if err != nil {
            return nil, err
        }

        gas = adj
    }

    taxes, err := FilterMsgAndComputeTax(clientCtx, msgs...)
    if err != nil {
        return nil, err
    }

    fees := txf.Fees().Add(taxes...)
    gasPrices := txf.GasPrices()

    if !gasPrices.IsZero() {
        glDec := math.LegacyNewDec(int64(gas))
        adjustment := math.LegacyNewDecWithPrec(int64(txf.GasAdjustment())*100, 2)
        if adjustment.LT(math.LegacyOneDec()) {
            adjustment = math.LegacyOneDec()
        }

        gasFees := make(sdk.Coins, len(gasPrices))
        for i, gp := range gasPrices {
            fee := gp.Amount.Mul(glDec).Mul(adjustment)
            gasFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
        }

        fees = fees.Add(gasFees.Sort()...)
    }

    return &legacytx.StdFee{
        Amount: fees,
        Gas:    gas,
    }, nil
}

func FilterMsgAndComputeTax(clientCtx client.Context, msgs ...sdk.Msg) (sdk.Coins, error) {
    return sdk.NewCoins(), nil
}

func prepareFactory(clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
    from := clientCtx.GetFromAddress()
    if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
        return txf, err
    }

    initNum, initSeq := txf.AccountNumber(), txf.Sequence()
    if initNum == 0 || initSeq == 0 {
        num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
        if err != nil {
            return txf, err
        }

        if initNum == 0 {
            txf = txf.WithAccountNumber(num)
        }

        if initSeq == 0 {
            txf = txf.WithSequence(seq)
        }
    }

    return txf, nil
}
