package oracle

import (
	"time"

	"cosmossdk.io/math"
	core "github.com/Daviddochain/dochain-core/v4/types"
	"github.com/Daviddochain/dochain-core/v4/x/oracle/keeper"
	"github.com/Daviddochain/dochain-core/v4/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker is called at the end of every block.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if core.IsPeriodLastBlock(ctx, params.VotePeriod) {
		// Build claim map over all validators in the active set.
		validatorClaimMap := make(map[string]types.Claim)

		maxValidators, err := k.StakingKeeper.MaxValidators(ctx)
		if err != nil {
			return
		}

		iterator, err := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		if err != nil {
			return
		}
		defer iterator.Close()

		powerReduction := k.StakingKeeper.PowerReduction(ctx)

		i := 0
		for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
			validator, err := k.StakingKeeper.Validator(ctx, iterator.Value())
			if err != nil {
				continue
			}

			// Exclude non-bonded validators.
			if validator.IsBonded() {
				valAddrStr := validator.GetOperator()
				valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
				if err != nil {
					continue
				}

				validatorClaimMap[valAddrStr] = types.NewClaim(
					validator.GetConsensusPower(powerReduction),
					0,
					0,
					valAddr,
				)
				i++
			}
		}

		// Denom-to-tobin-tax map.
		voteTargets := make(map[string]math.LegacyDec)
		k.IterateTobinTaxes(ctx, func(denom string, tobinTax math.LegacyDec) bool {
			voteTargets[denom] = tobinTax
			return false
		})

		// Clear all exchange rates.
		k.IterateDoExchangeRates(ctx, func(denom string, _ math.LegacyDec) (stop bool) {
			k.DeleteDoExchangeRate(ctx, denom)
			return false
		})

		// Organize votes into ballots by denom.
		// NOTE: Filter out inactive or jailed validators.
		// NOTE: Abstain votes have zero voting power.
		voteMap := k.OrganizeBallotByDenom(ctx, validatorClaimMap)

		if referenceDo := PickReferenceDo(ctx, k, voteTargets, voteMap); referenceDo != "" {
			// Build a vote map for the reference denom to calculate cross exchange rates.
			ballotRT := voteMap[referenceDo]
			voteMapRT := ballotRT.ToMap()
			exchangeRateRT := ballotRT.WeightedMedian()

			// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
			for denom, ballot := range voteMap {
				// Convert ballot to cross exchange rates.
				if denom != referenceDo {
					ballot = ballot.ToCrossRateWithSort(voteMapRT)
				}

				// Get weighted median of cross exchange rates.
				exchangeRate := Tally(ballot, params.RewardBand, validatorClaimMap)

				// Transform back into the original Do/asset quote form.
				if denom != referenceDo {
					exchangeRate = exchangeRateRT.Quo(exchangeRate)
				}

				// Set the exchange rate and emit the ABCI event.
				k.SetDoExchangeRateWithEvent(ctx, denom, exchangeRate)
			}
		}

		// ---------------------------
		// Do miss counting & slashing
		voteTargetsLen := len(voteTargets)
		for _, claim := range validatorClaimMap {
			// Skip abstain and valid voters.
			if int(claim.WinCount) == voteTargetsLen {
				continue
			}

			// Increase miss counter.
			k.SetMissCounter(ctx, claim.Recipient, k.GetMissCounter(ctx, claim.Recipient)+1)
		}

		// Distribute rewards to ballot winners.
		k.RewardBallotWinners(
			ctx,
			int64(params.VotePeriod),
			int64(params.RewardDistributionWindow),
			voteTargets,
			validatorClaimMap,
		)

		// Clear the ballot.
		k.ClearBallots(ctx, params.VotePeriod)

		// Update vote targets and tobin tax.
		k.ApplyWhitelist(ctx, params.Whitelist, voteTargets)
	}

	// Slash validators who missed voting over the threshold and
	// reset miss counters of all validators at the last block of the slash window.
	if core.IsPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}
}