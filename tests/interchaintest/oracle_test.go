package interchaintest

import (
	"context"
	"fmt"
	"sync"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/Daviddochain/dochain-core/v4/test/interchaintest/helpers"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestOracle(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	numVals := 3
	numFullNodes := 3

	config, err := createConfig()
	require.NoError(t, err)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "do",
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	do := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().AddChain(do)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	require.NoError(t, testutil.WaitForBlocks(ctx, 1, do))

	// Fund for 8 users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(genesisWalletAmount), do, do, do, do, do, do, do, do, do)

	require.NoError(t, testutil.WaitForBlocks(ctx, 5, do))

	height1, err := do.Height(ctx)
	require.NoError(t, err)

	// Create error channels for operations
	oracleErrCh := make(chan error, len(do.Validators))
	var wg sync.WaitGroup

	wg.Add(len(do.Validators))
	for _, val := range do.Validators {
		val := val
		go func(validator *cosmos.ChainNode) {
			defer wg.Done()
			for i := 0; i < 6; i++ {
				if err := helpers.ExecOracleMsgAggragatePrevote(ctx, validator, "salt", "1.123uusd"); err != nil {
					oracleErrCh <- err
					return
				}
				if err := testutil.WaitForBlocks(ctx, 1, do); err != nil {
					oracleErrCh <- err
					return
				}
			}
		}(val)
	}

	wg.Add(len(users))
	for i := range users {
		i := i
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				err := do.SendFunds(ctx, users[i].KeyName(), ibc.WalletAmount{
					Address: users[0].FormattedAddress(),
					Denom:   do.Config().Denom,
					Amount:  sdkmath.OneInt(),
				})
				require.NoError(t, err)
				require.NoError(t, testutil.WaitForBlocks(ctx, 1, do))
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(oracleErrCh)

	// Check for any errors that occurred in oracle operations
	for err := range oracleErrCh {
		require.NoError(t, err)
	}

	height2, err := do.Height(ctx)
	require.NoError(t, err)

	for h := height1; h <= height2; h++ {
		txs, err := do.Validators[2].FindTxs(ctx, h)
		convertedTxs := make([]Tx, len(txs))
		for i, tx := range txs {
			convertedEvents := make([]Event, len(tx.Events))

			for j, event := range tx.Events {
				convertedEvents[j] = Event{
					Type:       event.Type,
					Attributes: make([]EventAttribute, len(event.Attributes)),
				}

				for k, attr := range event.Attributes {
					convertedEvents[j].Attributes[k] = EventAttribute{
						Key:   attr.Key,
						Value: attr.Value,
					}
				}
			}

			convertedTxs[i] = Tx{
				Data:   tx.Data,
				Events: convertedEvents,
			}
		}
		for i, tx := range convertedTxs {
			fmt.Println("Tx: ", i)
			for _, event := range tx.Events {
				fmt.Println(event.Attributes)
			}
		}

		if !isOraclePrioritized(convertedTxs) {
			fmt.Println("Oracle transactions are not prioritized")
		}
		require.NoError(t, err)
	}

	// // Verify final validator state
	// stdout, _, err := do.Validators[0].ExecQuery(ctx, "staking", "validators")
	// require.NoError(t, err)
	// require.NotEmpty(t, stdout)

	// doValidators, _, err := helpers.UnmarshalValidators(*config.EncodingConfig, stdout)
	// require.NoError(t, err)
	// require.Equal(t, len(doValidators), 3)
}

type Tx struct {
	// For Tendermint transactions, this should be encoded as JSON.
	// Otherwise, this should be a human-readable format if possible.
	Data []byte

	// Events associated with the transaction, if applicable.
	Events []Event
}

// Event is an alternative representation of tendermint/abci/types.Event,
// so that the blockdb package does not depend directly on tendermint.
type Event struct {
	Type       string
	Attributes []EventAttribute

	// Notably, not including the Index field from the tendermint event.
	// The ABCI docs state:
	//
	// "The index flag notifies the Tendermint indexer to index the attribute. The value of the index flag is non-deterministic and may vary across different nodes in the network."
}

type EventAttribute struct {
	Key, Value string
}

func isOraclePrioritized(tx []Tx) bool {
	if len(tx) == 0 {
		return true
	}
	lastOracleIdx := -1
	firstNonOracleIdx := -1
	for i, t := range tx {
		if isOracleTx(t) {
			lastOracleIdx = i
			if firstNonOracleIdx != -1 {
				break
			}
		} else if firstNonOracleIdx == -1 {
			firstNonOracleIdx = i
		}
	}
	return lastOracleIdx == -1 || lastOracleIdx < firstNonOracleIdx
}

func isOracleTx(tx Tx) bool {
	for _, event := range tx.Events {
		for _, attr := range event.Attributes {
			if attr.Key == "oracle" {
				return true
			}
		}
	}
	return false
}






