package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// Keeper of the oracle store
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramSpace paramstypes.Subspace

	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	distrKeeper   types.DistributionKeeper
	StakingKeeper types.StakingKeeper

	distrName string
}

// NewKeeper constructs a new keeper for oracle
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	paramspace paramstypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	distrKeeper types.DistributionKeeper,
	stakingKeeper types.StakingKeeper,
	distrName string,
) Keeper {
	// ensure oracle module account is set
	if addr := accountKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramspace.HasKeyTable() {
		paramspace = paramspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		paramSpace:    paramspace,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		distrKeeper:   distrKeeper,
		StakingKeeper: stakingKeeper,
		distrName:     distrName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

//-----------------------------------
// ExchangeRate logic

// GetDoExchangeRate gets the consensus exchange rate of Do denominated in the denom asset from the store.
func (k Keeper) GetDoExchangeRate(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	if denom == core.MicroDoDenom {
		return math.LegacyOneDec(), nil
	}

	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetExchangeRateKey(denom))
	if b == nil {
		return math.LegacyZeroDec(), errorsmod.Wrap(types.ErrUnknownDenom, denom)
	}

	dp := sdk.DecProto{}
	k.cdc.MustUnmarshal(b, &dp)
	return dp.Dec, nil
}

// SetDoExchangeRate sets the consensus exchange rate of Do denominated in the denom asset to the store.
func (k Keeper) SetDoExchangeRate(ctx sdk.Context, denom string, exchangeRate math.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: exchangeRate})
	store.Set(types.GetExchangeRateKey(denom), bz)
}

// SetDoExchangeRateWithEvent sets the consensus exchange rate of Do
// denominated in the denom asset to the store with ABCI event
func (k Keeper) SetDoExchangeRateWithEvent(ctx sdk.Context, denom string, exchangeRate math.LegacyDec) {
	k.SetDoExchangeRate(ctx, denom, exchangeRate)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(types.EventTypeExchangeRateUpdate,
			sdk.NewAttribute(types.AttributeKeyDenom, denom),
			sdk.NewAttribute(types.AttributeKeyExchangeRate, exchangeRate.String()),
		),
	)
}

// DeleteDoExchangeRate deletes the consensus exchange rate of Do denominated in the denom asset from the store.
func (k Keeper) DeleteDoExchangeRate(ctx sdk.Context, denom string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetExchangeRateKey(denom))
}

// IterateDoExchangeRates iterates over do rates in the store
func (k Keeper) IterateDoExchangeRates(ctx sdk.Context, handler func(denom string, exchangeRate math.LegacyDec) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.ExchangeRateKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom := string(iter.Key()[len(types.ExchangeRateKey):])
		dp := sdk.DecProto{}
		k.cdc.MustUnmarshal(iter.Value(), &dp)
		if handler(denom, dp.Dec) {
			break
		}
	}
}

//-----------------------------------
// Oracle delegation logic

// GetFeederDelegation gets the account address that the validator operator delegated oracle vote rights to
func (k Keeper) GetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress) sdk.AccAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFeederDelegationKey(operator))
	if bz == nil {
		// By default the right is delegated to the validator itself
		return sdk.AccAddress(operator)
	}

	return sdk.AccAddress(bz)
}

// SetFeederDelegation sets the account address that the validator operator delegated oracle vote rights to
func (k Keeper) SetFeederDelegation(ctx sdk.Context, operator sdk.ValAddress, delegatedFeeder sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFeederDelegationKey(operator), delegatedFeeder.Bytes())
}

// IterateFeederDelegations iterates over the feed delegates and performs a callback function.
func (k Keeper) IterateFeederDelegations(ctx sdk.Context,
	handler func(delegator sdk.ValAddress, delegate sdk.AccAddress) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.FeederDelegationKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		delegator := sdk.ValAddress(iter.Key()[2:])
		delegate := sdk.AccAddress(iter.Value())

		if handler(delegator, delegate) {
			break
		}
	}
}

//-----------------------------------
// Miss counter logic

// GetMissCounter retrieves the # of vote periods missed in this oracle slash window
func (k Keeper) GetMissCounter(ctx sdk.Context, operator sdk.ValAddress) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMissCounterKey(operator))
	if bz == nil {
		// By default the counter is zero
		return 0
	}

	var missCounter gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &missCounter)
	return missCounter.Value
}

// SetMissCounter updates the # of vote periods missed in this oracle slash window
func (k Keeper) SetMissCounter(ctx sdk.Context, operator sdk.ValAddress, missCounter uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: missCounter})
	store.Set(types.GetMissCounterKey(operator), bz)
}

// DeleteMissCounter removes miss counter for the validator
func (k Keeper) DeleteMissCounter(ctx sdk.Context, operator sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetMissCounterKey(operator))
}

// IterateMissCounters iterates over the miss counters and performs a callback function.
func (k Keeper) IterateMissCounters(ctx sdk.Context,
	handler func(operator sdk.ValAddress, missCounter uint64) (stop bool),
) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.MissCounterKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		operator := sdk.ValAddress(iter.Key()[2:])

		var missCounter gogotypes.UInt64Value
		k.cdc.MustUnmarshal(iter.Value(), &missCounter)

		if handler(operator, missCounter.Value) {
			break
		}
	}
}

//-----------------------------------
// AggregateDoRatePrevote logic

// GetAggregateDoRatePrevote retrieves an oracle prevote from the store
func (k Keeper) GetAggregateDoRatePrevote(ctx sdk.Context, voter sdk.ValAddress) (aggregatePrevote types.AggregateDoRatePrevote, err error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetAggregateDoRatePrevoteKey(voter))
	if b == nil {
		err = errorsmod.Wrap(types.ErrNoAggregatePrevote, voter.String())
		return aggregatePrevote, err
	}
	k.cdc.MustUnmarshal(b, &aggregatePrevote)
	return aggregatePrevote, err
}

// SetAggregateDoRatePrevote set an oracle aggregate prevote to the store
func (k Keeper) SetAggregateDoRatePrevote(ctx sdk.Context, voter sdk.ValAddress, prevote types.AggregateDoRatePrevote) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&prevote)

	store.Set(types.GetAggregateDoRatePrevoteKey(voter), bz)
}

// DeleteAggregateDoRatePrevote deletes an oracle prevote from the store
func (k Keeper) DeleteAggregateDoRatePrevote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateDoRatePrevoteKey(voter))
}

// IterateAggregateDoRatePrevotes iterates rate over prevotes in the store
func (k Keeper) IterateAggregateDoRatePrevotes(ctx sdk.Context, handler func(voterAddr sdk.ValAddress, aggregatePrevote types.AggregateDoRatePrevote) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.AggregateDoRatePrevoteKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregatePrevote types.AggregateDoRatePrevote
		k.cdc.MustUnmarshal(iter.Value(), &aggregatePrevote)
		if handler(voterAddr, aggregatePrevote) {
			break
		}
	}
}

//-----------------------------------
// AggregateDoRateVote logic

// GetAggregateDoRateVote retrieves an oracle prevote from the store
func (k Keeper) GetAggregateDoRateVote(ctx sdk.Context, voter sdk.ValAddress) (aggregateVote types.AggregateDoRateVote, err error) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.GetAggregateDoRateVoteKey(voter))
	if b == nil {
		err = errorsmod.Wrap(types.ErrNoAggregateVote, voter.String())
		return aggregateVote, err
	}
	k.cdc.MustUnmarshal(b, &aggregateVote)
	return aggregateVote, err
}

// SetAggregateDoRateVote adds an oracle aggregate prevote to the store
func (k Keeper) SetAggregateDoRateVote(ctx sdk.Context, voter sdk.ValAddress, vote types.AggregateDoRateVote) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&vote)
	store.Set(types.GetAggregateDoRateVoteKey(voter), bz)
}

// DeleteAggregateDoRateVote deletes an oracle prevote from the store
func (k Keeper) DeleteAggregateDoRateVote(ctx sdk.Context, voter sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAggregateDoRateVoteKey(voter))
}

// IterateAggregateDoRateVotes iterates rate over prevotes in the store
func (k Keeper) IterateAggregateDoRateVotes(ctx sdk.Context, handler func(voterAddr sdk.ValAddress, aggregateVote types.AggregateDoRateVote) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.AggregateDoRateVoteKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		voterAddr := sdk.ValAddress(iter.Key()[2:])

		var aggregateVote types.AggregateDoRateVote
		k.cdc.MustUnmarshal(iter.Value(), &aggregateVote)
		if handler(voterAddr, aggregateVote) {
			break
		}
	}
}

// GetTobinTax return tobin tax for the denom
func (k Keeper) GetTobinTax(ctx sdk.Context, denom string) (math.LegacyDec, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTobinTaxKey(denom))
	if bz == nil {
		err := errorsmod.Wrap(types.ErrNoTobinTax, denom)
		return math.LegacyDec{}, err
	}

	tobinTax := sdk.DecProto{}
	k.cdc.MustUnmarshal(bz, &tobinTax)

	return tobinTax.Dec, nil
}

// SetTobinTax updates tobin tax for the denom
func (k Keeper) SetTobinTax(ctx sdk.Context, denom string, tobinTax math.LegacyDec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&sdk.DecProto{Dec: tobinTax})
	store.Set(types.GetTobinTaxKey(denom), bz)
}

// IterateTobinTaxes iterates rate over tobin taxes in the store
func (k Keeper) IterateTobinTaxes(ctx sdk.Context, handler func(denom string, tobinTax math.LegacyDec) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.TobinTaxKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		denom := types.ExtractDenomFromTobinTaxKey(iter.Key())

		var tobinTax sdk.DecProto
		k.cdc.MustUnmarshal(iter.Value(), &tobinTax)
		if handler(denom, tobinTax.Dec) {
			break
		}
	}
}

// ClearTobinTaxes clears tobin taxes
func (k Keeper) ClearTobinTaxes(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	iter := storetypes.KVStorePrefixIterator(store, types.TobinTaxKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
}

// ValidateFeeder return the given feeder is allowed to feed the message or not
func (k Keeper) ValidateFeeder(ctx sdk.Context, feederAddr sdk.AccAddress, validatorAddr sdk.ValAddress) error {
	if !feederAddr.Equals(validatorAddr) {
		delegate := k.GetFeederDelegation(ctx, validatorAddr)
		if !delegate.Equals(feederAddr) {
			return errorsmod.Wrap(types.ErrNoVotingPermission, feederAddr.String())
		}
	}

	// Check that the given validator exists
	if val, err := k.StakingKeeper.Validator(ctx, validatorAddr); val == nil || !val.IsBonded() || err != nil {
		if err != nil {
			return err
		}
		return errorsmod.Wrapf(stakingtypes.ErrNoValidatorFound, "validator %s is not active set", validatorAddr.String())
	}

	return nil
}






