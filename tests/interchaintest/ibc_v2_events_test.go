package interchaintest

import (
	"context"
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// Local IBC event types (avoid depending on other test files)
type IBCEventTx struct {
	Data   []byte
	Events []IBCEvent
}

type IBCEvent struct {
	Type       string
	Attributes []IBCEventAttribute
}

type IBCEventAttribute struct {
	Key   string
	Value string
}

// scanBlockEvents converts txs for a given height from a node into the local Tx representation used in oracle_test.go
func scanBlockEvents(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, height int64) []IBCEventTx {
	n := chain.Validators[0]
	txs, err := n.FindTxs(ctx, height)
	require.NoError(t, err)
	convertedTxs := make([]IBCEventTx, len(txs))
	for i, tx := range txs {
		convertedEvents := make([]IBCEvent, len(tx.Events))
		for j, event := range tx.Events {
			convertedEvents[j] = IBCEvent{
				Type:       event.Type,
				Attributes: make([]IBCEventAttribute, len(event.Attributes)),
			}
			for k, attr := range event.Attributes {
				convertedEvents[j].Attributes[k] = IBCEventAttribute{Key: attr.Key, Value: attr.Value}
			}
		}
		convertedTxs[i] = IBCEventTx{Data: tx.Data, Events: convertedEvents}
	}
	return convertedTxs
}

func containsEvent(txs []IBCEventTx, eventType string) bool {
	for _, tx := range txs {
		for _, ev := range tx.Events {
			if strings.EqualFold(ev.Type, eventType) { // be lenient on case
				return true
			}
		}
	}
	return false
}

func containsAnyEventInWindow(t *testing.T, ctx context.Context, chain *cosmos.CosmosChain, startHeight, endHeight int64, eventType string) bool {
	for h := startHeight; h <= endHeight; h++ {
		txs := scanBlockEvents(t, ctx, chain, h)
		if containsEvent(txs, eventType) {
			return true
		}
	}
	return false
}

// TestIBCv2HandshakeEvents validates that channel and connection handshake events are emitted during path creation.
func TestIBCv2HandshakeEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	// Do not run in parallel; consumes resources shared with other tests

	numVals := 3
	numFullNodes := 3

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
			ChainConfig:   createGaiaConfig(),
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	do, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)

	const path = "ibcv2-handshake"

	ic := interchaintest.NewInterchain().
		AddChain(do).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{Chain1: do, Chain2: gaia, Relayer: r, Path: path})

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{TestName: t.Name(), Client: client, NetworkID: network}))
	t.Cleanup(func() { _ = ic.Close() })

	// After build, handshake should have completed. Capture current heights and scan a recent window.
	require.NoError(t, testutil.WaitForBlocks(ctx, 3, do, gaia))
	doH, err := do.Height(ctx)
	require.NoError(t, err)
	gaiaH, err := gaia.Height(ctx)
	require.NoError(t, err)

	startDo Chain := doH - 50
	if startDo Chain < 1 {
		startDo Chain = 1
	}
	startGaia := gaiaH - 50
	if startGaia < 1 {
		startGaia = 1
	}

	// Expected IBC v2 handshake events
	handshakeEvents := []string{
		"connection_open_init",
		"connection_open_try",
		"connection_open_ack",
		"connection_open_confirm",
		"channel_open_init",
		"channel_open_try",
		"channel_open_ack",
		"channel_open_confirm",
	}

	for _, ev := range handshakeEvents {
		found := containsAnyEventInWindow(t, ctx, do, startDo Chain, doH, ev) ||
			containsAnyEventInWindow(t, ctx, gaia, startGaia, gaiaH, ev)
		require.Truef(t, found, "expected to find event %s in recent handshake window", ev)
	}
}

// TestIBCv2TransferEvents validates send, recv, ack events for a standard ICS20 transfer
func TestIBCv2TransferEvents(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	numVals := 3
	numFullNodes := 3

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
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
			ChainConfig:   createGaiaConfig(),
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	do, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).Build(t, client, network)
	ic := interchaintest.NewInterchain().
		AddChain(do).
		AddChain(gaia).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{Chain1: do, Chain2: gaia, Relayer: r, Path: pathDoGaia})

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{TestName: t.Name(), Client: client, NetworkID: network}))
	t.Cleanup(func() { _ = ic.Close() })

	require.NoError(t, r.StartRelayer(ctx, eRep, pathDoGaia))
	t.Cleanup(func() { _ = r.StopRelayer(ctx, eRep) })

	// Fund users
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.NewInt(genesisWalletAmount), do, gaia)
	doUser, gaiaUser := users[0], users[1]
	require.NoError(t, testutil.WaitForBlocks(ctx, 5, do, gaia))

	channel, err := ibc.GetTransferChannel(ctx, r, eRep, do.Config().ChainID, gaia.Config().ChainID)
	require.NoError(t, err)

	transfer := ibc.WalletAmount{Address: gaiaUser.FormattedAddress(), Denom: do.Config().Denom, Amount: math.NewInt(1000)}
	transferTx, err := do.SendIBCTransfer(ctx, channel.ChannelID, doUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	doH, err := do.Height(ctx)
	require.NoError(t, err)

	_, err = testutil.PollForAck(ctx, do, doH-5, doH+25, transferTx.Packet)
	require.NoError(t, err)
	// give relayer time to relay ack
	require.NoError(t, testutil.WaitForBlocks(ctx, 3, do, gaia))

	// Scan recent window for events
	doH2, _ := do.Height(ctx)
	gaiaH2, _ := gaia.Height(ctx)
	startDo Chain := doH - 10
	if startDo Chain < 1 {
		startDo Chain = 1
	}
	startGaia := gaiaH2 - 30
	if startGaia < 1 {
		startGaia = 1
	}

	require.True(t, containsAnyEventInWindow(t, ctx, do, startDo Chain, doH2, "send_packet"))
	// recv and write_ack occur on destination
	require.True(t, containsAnyEventInWindow(t, ctx, gaia, startGaia, gaiaH2, "recv_packet"))
	require.True(t, containsAnyEventInWindow(t, ctx, gaia, startGaia, gaiaH2, "write_acknowledgement"))
	// acknowledge_packet occurs on source when ack is relayed back
	require.True(t, containsAnyEventInWindow(t, ctx, do, startDo Chain, doH2, "acknowledge_packet"))
}






