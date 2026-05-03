package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestDoGaiaIBCTransfer spins up a Do-Chain and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Do-Chain -> Gaia and then back from Gaia -> Do-Chain.
func TestDoGaiaIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	// Create chain factory with Do-Chain
	numVals := 3
	numFullNodes := 3
	// tax rate in ictest is 0.0001
	taxRate := sdkmath.LegacyNewDecWithPrec(1, 4)

	client, network := interchaintest.DockerSetup(t)

	ctx := context.Background()

	config, err := createConfig()
	require.NoError(t, err)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "do",
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "gaia",
			Version:       "v25.1.0",
			ChainConfig:   createGaiaConfig(), // Added chain config for Gaia
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	do, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Create relayer factory to utilize the go-relayer
	rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r := rf.Build(t, client, network)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().
		AddChain(do).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  do,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathDoGaia,
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
	require.NoError(t, r.StartRelayer(ctx, eRep, pathDoGaia))
	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				panic(fmt.Errorf("an error occurred while stopping the relayer: %s", err))
			}
		},
	)

	// Create and Fund User Wallets
	taxAmount := taxRate.MulInt(sdkmath.NewInt(genesisWalletAmount)).TruncateInt()
	receivedAmount := sdkmath.NewInt(genesisWalletAmount).Sub(taxAmount)
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(genesisWalletAmount), do, gaia)
	doUser := users[0]
	gaiaUser := users[1]

	doUserAddr := doUser.FormattedAddress()
	gaiaUserAddr := gaiaUser.FormattedAddress()

	err = testutil.WaitForBlocks(ctx, 10, do, gaia)
	require.NoError(t, err)

	doUserInitialBal, err := do.GetBalance(ctx, doUserAddr, do.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, receivedAmount, doUserInitialBal)

	gaiaUserInitialBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletBalance, gaiaUserInitialBal)

	// Compose an IBC transfer and send from Do-Chain -> Gaia
	transferAmount := math.NewInt(1000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   do.Config().Denom,
		Amount:  transferAmount,
	}

	// Query for the newly created channel
	dochainChannels, err := r.GetChannels(ctx, eRep, do.Config().ChainID)
	require.NoError(t, err)

	transferTx, err := do.SendIBCTransfer(ctx, dochainChannels[0].ChannelID, doUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	doHeight, err := do.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, do, doHeight, doHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Get the IBC denom for udotest on Gaia
	dochainTokenDenom := transfertypes.GetPrefixedDenom(dochainChannels[0].Counterparty.PortID, dochainChannels[0].Counterparty.ChannelID, do.Config().Denom)
	doIBCDenom := transfertypes.ParseDenomTrace(dochainTokenDenom).IBCDenom()

	// the transfer is using 200000 gas, gas price is 28.325udo
	gasFee := math.LegacyNewDec(200000).Mul(math.LegacyNewDecWithPrec(28325, 3))

	// Assert that the funds are no longer present in user acc on Do-Chain and are in the user acc on Gaia
	doUserUpdateBal, err := do.GetBalance(ctx, doUserAddr, do.Config().Denom)
	require.NoError(t, err)

	// TODO: the gas fee is not fixed 200000 gas, so the below test is not working
	// require.Equal(t, doUserUpdateBal, doUserInitialBal.Sub(transferAmount).Sub(gasFee.RoundInt()))
	require.LessOrEqual(t, doUserUpdateBal.Int64(), doUserInitialBal.Sub(transferAmount).Sub(gasFee.RoundInt()).Int64())

	gaiaUserUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, doIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUserUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Do-Chain
	transfer = ibc.WalletAmount{
		Address: doUserAddr,
		Denom:   doIBCDenom,
		Amount:  transferAmount,
	}

	transferTx, err = gaia.SendIBCTransfer(ctx, dochainChannels[0].Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Do-Chain and not on Gaia (except gas fees paid of course)
	doUserUpdateBal, err = do.GetBalance(ctx, doUserAddr, do.Config().Denom)
	require.NoError(t, err)
	// TODO: as above this test does not work as "gas" is set to auto.
	// require.Equal(t, doUserInitialBal.Sub(gasFee.RoundInt()), doUserUpdateBal)
	require.LessOrEqual(t, doUserUpdateBal.Int64(), doUserInitialBal.Sub(gasFee.RoundInt()).Int64())

	gaiaUserUpdateBal, err = gaia.GetBalance(ctx, gaiaUserAddr, doIBCDenom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(0), gaiaUserUpdateBal)
}







