package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/market/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ApplySwapToPool updates each pool with offerCoin and askCoin taken from a swap operation.
// OfferPool = OfferPool + offerAmt
// AskPool   = AskPool - askAmt
func (k Keeper) ApplySwapToPool(ctx sdk.Context, offerCoin sdk.Coin, askCoin sdk.DecCoin) error {
	// No delta update in case of Do to Do swap.
	if offerCoin.Denom != core.MicroDoDenom && askCoin.Denom != core.MicroDoDenom {
		return nil
	}

	doPoolDelta := k.GetDoPoolDelta(ctx)

	// In case of swapping a non-Do asset to Do, the Do swap pool must be increased.
	if offerCoin.Denom != core.MicroDoDenom && askCoin.Denom == core.MicroDoDenom {
		offerBaseCoin, err := k.ComputeInternalSwap(ctx, sdk.NewDecCoinFromCoin(offerCoin), core.MicroDoDenom)
		if err != nil {
			return err
		}

		doPoolDelta = doPoolDelta.Add(offerBaseCoin.Amount)
	}

	// In case of swapping Do to a non-Do asset, the Do swap pool must be decreased.
	if offerCoin.Denom == core.MicroDoDenom && askCoin.Denom != core.MicroDoDenom {
		askBaseCoin, err := k.ComputeInternalSwap(ctx, askCoin, core.MicroDoDenom)
		if err != nil {
			return err
		}

		doPoolDelta = doPoolDelta.Sub(askBaseCoin.Amount)
	}

	k.SetDoPoolDelta(ctx, doPoolDelta)

	return nil
}

// ComputeSwap returns the amount of asked coins that should be returned for a given
// offerCoin at the effective exchange rate registered with the oracle.
// Returns an error if the swap is recursive, or the coins to be traded are unknown
// by the oracle, or the amount to trade is too small.
func (k Keeper) ComputeSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (retDecCoin sdk.DecCoin, spread math.LegacyDec, err error) {
	// Return invalid recursive swap error.
	if offerCoin.Denom == askDenom {
		return sdk.DecCoin{}, math.LegacyZeroDec(), errorsmod.Wrap(types.ErrRecursiveSwap, askDenom)
	}

	// Swap the offer coin to the chain base denom for simplicity of the swap process.
	baseOfferDecCoin, err := k.ComputeInternalSwap(ctx, sdk.NewDecCoinFromCoin(offerCoin), core.MicroDoDenom)
	if err != nil {
		return sdk.DecCoin{}, math.LegacyDec{}, err
	}

	// Get swap amount based on the oracle price.
	retDecCoin, err = k.ComputeInternalSwap(ctx, baseOfferDecCoin, askDenom)
	if err != nil {
		return sdk.DecCoin{}, math.LegacyDec{}, err
	}

	// Non-Do to non-Do swap: apply only tobin tax without constant-product spread.
	if offerCoin.Denom != core.MicroDoDenom && askDenom != core.MicroDoDenom {
		var tobinTax math.LegacyDec

		offerTobinTax, err2 := k.OracleKeeper.GetTobinTax(ctx, offerCoin.Denom)
		if err2 != nil {
			return sdk.DecCoin{}, math.LegacyDec{}, err2
		}

		askTobinTax, err2 := k.OracleKeeper.GetTobinTax(ctx, askDenom)
		if err2 != nil {
			return sdk.DecCoin{}, math.LegacyDec{}, err2
		}

		// Apply the higher tobin tax of the two denoms in the swap operation.
		if askTobinTax.GT(offerTobinTax) {
			tobinTax = askTobinTax
		} else {
			tobinTax = offerTobinTax
		}

		spread = tobinTax
		return retDecCoin, spread, nil
	}

	basePool := k.BasePool(ctx)
	minSpread := k.MinStabilitySpread(ctx)

	// Constant-product pool, which by construction is the square of the equilibrium base pool.
	cp := basePool.Mul(basePool)
	doPoolDelta := k.GetDoPoolDelta(ctx)
	doPool := basePool.Add(doPoolDelta)
	baseAssetPool := cp.Quo(doPool)

	var offerPool math.LegacyDec // base denom (udo) unit
	var askPool math.LegacyDec   // base denom (udo) unit

	if offerCoin.Denom != core.MicroDoDenom {
		// non-Do -> Do swap
		offerPool = doPool
		askPool = baseAssetPool
	} else {
		// Do -> non-Do swap
		offerPool = baseAssetPool
		askPool = doPool
	}

	// Get constant-product based swap amount:
	// askBaseAmount = askPool - cp / (offerPool + offerBaseAmount)
	// askBaseAmount is in base denom (udo) units.
	askBaseAmount := askPool.Sub(cp.Quo(offerPool.Add(baseOfferDecCoin.Amount)))

	// Both baseOffer and baseAsk are base-denom units, so spread can be calculated by:
	// spread = (baseOfferAmt - baseAskAmt) / baseOfferAmt
	baseOfferAmount := baseOfferDecCoin.Amount
	spread = baseOfferAmount.Sub(askBaseAmount).Quo(baseOfferAmount)

	if spread.LT(minSpread) {
		spread = minSpread
	}

	return retDecCoin, spread, nil
}

// ComputeInternalSwap returns the amount of asked DecCoin that should be returned
// for a given offerCoin at the effective exchange rate registered with the oracle.
// Unlike ComputeSwap, ComputeInternalSwap does not charge a spread because its use
// is internal to the module.
func (k Keeper) ComputeInternalSwap(ctx sdk.Context, offerCoin sdk.DecCoin, askDenom string) (sdk.DecCoin, error) {
	if offerCoin.Denom == askDenom {
		return offerCoin, nil
	}

	offerRate, err := k.OracleKeeper.GetDoExchangeRate(ctx, offerCoin.Denom)
	if err != nil {
		return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, offerCoin.Denom)
	}

	askRate, err := k.OracleKeeper.GetDoExchangeRate(ctx, askDenom)
	if err != nil {
		return sdk.DecCoin{}, errorsmod.Wrap(types.ErrNoEffectivePrice, askDenom)
	}

	retAmount := offerCoin.Amount.Mul(askRate).Quo(offerRate)
	if retAmount.LTE(math.LegacyZeroDec()) {
		return sdk.DecCoin{}, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, offerCoin.String())
	}

	return sdk.NewDecCoinFromDec(askDenom, retAmount), nil
}

// simulateSwap performs a simulated swap.
func (k Keeper) simulateSwap(ctx sdk.Context, offerCoin sdk.Coin, askDenom string) (sdk.Coin, error) {
	if askDenom == offerCoin.Denom {
		return sdk.Coin{}, errorsmod.Wrap(types.ErrRecursiveSwap, askDenom)
	}

	if offerCoin.Amount.BigInt().BitLen() > 100 {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrInvalidCoins, offerCoin.String())
	}

	swapCoin, spread, err := k.ComputeSwap(ctx, offerCoin, askDenom)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(sdkerrors.ErrPanic, err.Error())
	}

	if spread.IsPositive() {
		swapFeeAmt := spread.Mul(swapCoin.Amount)
		if swapFeeAmt.IsPositive() {
			swapFee := sdk.NewDecCoinFromDec(swapCoin.Denom, swapFeeAmt)
			swapCoin = swapCoin.Sub(swapFee)
		}
	}

	retCoin, _ := swapCoin.TruncateDecimal()
	return retCoin, nil
}