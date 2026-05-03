package interchaintest

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/Daviddochain/dochain-core/v4/test/interchaintest/helpers"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestIBCHooks ensures the ibc-hooks middleware from osmosis works.
func TestDoIBCHooks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Do-Chain
	numVals := 3
	numFullNodes := 3

	client, network := interchaintest.DockerSetup(t)

	ctx := context.Background()

	config1, err := createConfig()
	require.NoError(t, err)

	config2 := config1.Clone()
	config2.Name = "core-counterparty"
	config2.ChainID = "core-counterparty-1"

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "do",
			ChainConfig:   config1,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "do",
			ChainConfig:   config2,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	const (
		path = "ibc-path"
	)

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	do, dochain2 := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Create relayer factory to utilize the go-relayer
	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).
		Build(t, client, network)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().
		AddChain(do).
		AddChain(dochain2).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  do,
			Chain2:  dochain2,
			Relayer: r,
			Path:    path,
		})

	// Build interchain
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:  t.Name(),
		Client:    client,
		NetworkID: network,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Start the relayer and set the cleanup function.
	require.NoError(t, r.StartRelayer(ctx, eRep, path))
	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	// Create and Fund User Wallets
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(genesisWalletAmount), do, dochain2)
	doUser, dochain2User := users[0], users[1]

	doUserAddr := doUser.FormattedAddress()

	// Wait a few blocks for relayer to start and for user accounts to be created
	err = testutil.WaitForBlocks(ctx, 5, do, dochain2)
	require.NoError(t, err)

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, do.Config().ChainID, dochain2.Config().ChainID)
	require.NoError(t, err)

	_, contractAddr := helpers.SetupContract(t, ctx, dochain2, dochain2User.KeyName(), "bytecode/counter.wasm", `{"count":0}`)

	transfer := ibc.WalletAmount{
		Address: contractAddr,
		Denom:   do.Config().Denom,
		Amount:  math.OneInt(),
	}

	memo := ibc.TransferOptions{
		Memo: fmt.Sprintf(`{"wasm":{"contract":"%s","msg":%s}}`, contractAddr, `{"increment":{}}`),
	}

	// Initial transfer. Account is created by the wasm execute is not so we must do this twice to properly set up
	transferTx, err := do.SendIBCTransfer(ctx, channel.ChannelID, doUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	doHeight, err := do.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, do, doHeight-5, doHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Second time, this will make the counter == 1 since the account is now created.
	transferTx, err = do.SendIBCTransfer(ctx, channel.ChannelID, doUser.KeyName(), transfer, memo)
	require.NoError(t, err)
	doHeight, err = do.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, do, doHeight-5, doHeight+25, transferTx.Packet)
	require.NoError(t, err)

	// Get the address on the other chain's side
	addr := helpers.GetIBCHooksUserAddress(t, ctx, do, channel.ChannelID, doUserAddr)
	require.NotEmpty(t, addr)

	// Get funds on the receiving chain
	funds := helpers.GetIBCHookTotalFunds(t, ctx, dochain2, contractAddr, addr)
	require.Equal(t, int(1), len(funds.Data.TotalFunds))

	var ibcDenom string
	for _, coin := range funds.Data.TotalFunds {
		if strings.HasPrefix(coin.Denom, "ibc/") {
			ibcDenom = coin.Denom
			break
		}
	}
	require.NotEmpty(t, ibcDenom)

	channelsDo Chain, err := r.GetChannels(ctx, eRep, do.Config().ChainID)
	require.NoError(t, err)
	channelDo ChainDo Chain2 := channelsDo Chain[0]
	require.NotEmpty(t, channelDo ChainDo Chain2.ChannelID)

	dochainOnDo Chain2TokenDenom := transfertypes.GetPrefixedDenom(channelDo ChainDo Chain2.Counterparty.PortID, channelDo ChainDo Chain2.Counterparty.ChannelID, do.Config().Denom)
	dochainOnDo Chain2IBCDenom := transfertypes.ParseDenomTrace(dochainOnDo Chain2TokenDenom).IBCDenom()
	require.Equal(t, ibcDenom, dochainOnDo Chain2IBCDenom)

	// ensure the count also increased to 1 as expected.
	count := helpers.GetIBCHookCount(t, ctx, dochain2, contractAddr, addr)
	require.Equal(t, int64(1), count.Data.Count)
}






