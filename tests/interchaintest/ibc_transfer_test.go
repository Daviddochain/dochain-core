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

// TestTerraGaiaIBCTranfer spins up a Do-Chain and Gaia network, initializes an IBC connection between them,
// and sends an ICS20 token transfer from Do-Chain -> Gaia and then back from Gaia -> Do-Chain.
func TestTerraGaiaIBCTranfer(t *testing.T) {
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
			Name:          "dochain",
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

	dochain, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	// Create relayer factory to utilize the go-relayer
	rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t))
	r := rf.Build(t, client, network)

	// Create a new Interchain object which describes the chains, relayers, and IBC connections we want to use
	ic := interchaintest.NewInterchain().
		AddChain(dochain).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  dochain,
			Chain2:  gaia,
			Relayer: r,
			Path:    pathTerraGaia,
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
	require.NoError(t, r.StartRelayer(ctx, eRep, pathTerraGaia))
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
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", sdkmath.NewInt(genesisWalletAmount), dochain, gaia)
	terraUser := users[0]
	gaiaUser := users[1]

	terraUserAddr := terraUser.FormattedAddress()
	gaiaUserAddr := gaiaUser.FormattedAddress()

	err = testutil.WaitForBlocks(ctx, 10, dochain, gaia)
	require.NoError(t, err)

	terraUserInitialBal, err := dochain.GetBalance(ctx, terraUserAddr, dochain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, receivedAmount, terraUserInitialBal)

	gaiaUserInitialBal, err := gaia.GetBalance(ctx, gaiaUserAddr, gaia.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, genesisWalletBalance, gaiaUserInitialBal)

	// Compose an IBC transfer and send from Do-Chain -> Gaia
	transferAmount := math.NewInt(1000)
	transfer := ibc.WalletAmount{
		Address: gaiaUserAddr,
		Denom:   dochain.Config().Denom,
		Amount:  transferAmount,
	}

	// Query for the newly created channel
	terraChannels, err := r.GetChannels(ctx, eRep, dochain.Config().ChainID)
	require.NoError(t, err)

	transferTx, err := dochain.SendIBCTransfer(ctx, terraChannels[0].ChannelID, terraUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	terraHeight, err := dochain.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, dochain, terraHeight, terraHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Get the IBC denom for udo on Gaia
	terraTokenDenom := transfertypes.GetPrefixedDenom(terraChannels[0].Counterparty.PortID, terraChannels[0].Counterparty.ChannelID, dochain.Config().Denom)
	terraIBCDenom := transfertypes.ParseDenomTrace(terraTokenDenom).IBCDenom()

	// the transfer is using 200000 gas, gas price is 28.325uluna
	gasFee := math.LegacyNewDec(200000).Mul(math.LegacyNewDecWithPrec(28325, 3))

	// Assert that the funds are no longer present in user acc on Do-Chain and are in the user acc on Gaia
	terraUserUpdateBal, err := dochain.GetBalance(ctx, terraUserAddr, dochain.Config().Denom)
	require.NoError(t, err)

	// TODO: the gas fee is not fixed 200000 gas, so the below test is not working
	// require.Equal(t, terraUserUpdateBal, terraUserInitialBal.Sub(transferAmount).Sub(gasFee.RoundInt()))
	require.LessOrEqual(t, terraUserUpdateBal.Int64(), terraUserInitialBal.Sub(transferAmount).Sub(gasFee.RoundInt()).Int64())

	gaiaUserUpdateBal, err := gaia.GetBalance(ctx, gaiaUserAddr, terraIBCDenom)
	require.NoError(t, err)
	require.Equal(t, transferAmount, gaiaUserUpdateBal)

	// Compose an IBC transfer and send from Gaia -> Do-Chain
	transfer = ibc.WalletAmount{
		Address: terraUserAddr,
		Denom:   terraIBCDenom,
		Amount:  transferAmount,
	}

	transferTx, err = gaia.SendIBCTransfer(ctx, terraChannels[0].Counterparty.ChannelID, gaiaUserAddr, transfer, ibc.TransferOptions{})
	require.NoError(t, err)

	gaiaHeight, err := gaia.Height(ctx)
	require.NoError(t, err)

	// Poll for the ack to know the transfer was successful
	_, err = testutil.PollForAck(ctx, gaia, gaiaHeight, gaiaHeight+10, transferTx.Packet)
	require.NoError(t, err)

	// Assert that the funds are now back on Do-Chain and not on Gaia (except gas fees paid of course)
	terraUserUpdateBal, err = dochain.GetBalance(ctx, terraUserAddr, dochain.Config().Denom)
	require.NoError(t, err)
	// TODO: as above this test does not work as "gas" is set to auto.
	// require.Equal(t, terraUserInitialBal.Sub(gasFee.RoundInt()), terraUserUpdateBal)
	require.LessOrEqual(t, terraUserUpdateBal.Int64(), terraUserInitialBal.Sub(gasFee.RoundInt()).Int64())

	gaiaUserUpdateBal, err = gaia.GetBalance(ctx, gaiaUserAddr, terraIBCDenom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(0), gaiaUserUpdateBal)
}



