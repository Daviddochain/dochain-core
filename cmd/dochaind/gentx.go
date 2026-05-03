package main

import (
"bufio"
"bytes"
"encoding/json"
"fmt"
"io"
"os"
"path/filepath"

"cosmossdk.io/errors"
"github.com/spf13/cobra"

doapp "github.com/Daviddochain/dochain-core/v4/app"
"github.com/cosmos/cosmos-sdk/client"
"github.com/cosmos/cosmos-sdk/client/flags"
"github.com/cosmos/cosmos-sdk/client/tx"
"github.com/cosmos/cosmos-sdk/codec/address"
"github.com/cosmos/cosmos-sdk/crypto/keyring"
"github.com/cosmos/cosmos-sdk/server"
sdk "github.com/cosmos/cosmos-sdk/types"
"github.com/cosmos/cosmos-sdk/types/module"
"github.com/cosmos/cosmos-sdk/version"
authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
"github.com/cosmos/cosmos-sdk/x/genutil"
genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli"
stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func GenTxCmd(basicMgr module.BasicManager, txEncCfg client.TxEncodingConfig) *cobra.Command {
defaultNodeHome := doapp.DefaultNodeHome
genBalIterator := banktypes.GenesisBalancesIterator{}
valAddressCodec := address.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix())

ipDefault, _ := server.ExternalIP()
fsCreateValidator, defaultsDesc := stakingcli.CreateValidatorMsgFlagSet(ipDefault)

cmd := &cobra.Command{
Use:   "gentx [key_name] [amount]",
Short: "Generate a genesis tx carrying a self delegation",
Args:  cobra.ExactArgs(2),
Long: fmt.Sprintf(`Generate a genesis transaction that creates a validator with a self-delegation,
that is signed by the key in the Keyring referenced by a given name. A node ID and consensus
pubkey may optionally be provided. If they are omitted, they will be retrieved from the priv_validator.json
file. The following default parameters are included:

%s

Example:

$ %s gentx my-key-name 1000000stake --home=/path/to/home/dir --keyring-backend=os --chain-id=test-chain-1 \
  --moniker="myvalidator" \
  --commission-max-change-rate=0.01 \
  --commission-max-rate=1.0 \
  --commission-rate=0.07 \
  --details="..." \
  --security-contact="..." \
  --website="..."
`, defaultsDesc, version.AppName),
RunE: func(cmd *cobra.Command, args []string) error {
serverCtx := server.GetServerContextFromCmd(cmd)

clientCtx, err := client.GetClientTxContext(cmd)
if err != nil {
return err
}

cdc := clientCtx.Codec
config := serverCtx.Config
config.SetRoot(clientCtx.HomeDir)

nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(serverCtx.Config)
if err != nil {
return errors.Wrap(err, "failed to initialize node validator files")
}

if nodeIDString, _ := cmd.Flags().GetString(stakingcli.FlagNodeID); nodeIDString != "" {
nodeID = nodeIDString
}

if pkStr, _ := cmd.Flags().GetString(stakingcli.FlagPubKey); pkStr != "" {
if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkStr), &valPubKey); err != nil {
return errors.Wrap(err, "failed to unmarshal validator public key")
}
}

appGenesis, err := genutiltypes.AppGenesisFromFile(config.GenesisFile())
if err != nil {
return errors.Wrapf(err, "failed to read genesis doc file %s", config.GenesisFile())
}

var genesisState map[string]json.RawMessage
if err = json.Unmarshal(appGenesis.AppState, &genesisState); err != nil {
return errors.Wrap(err, "failed to unmarshal genesis state")
}

if err = basicMgr.ValidateGenesis(cdc, txEncCfg, genesisState); err != nil {
return errors.Wrap(err, "failed to validate genesis state")
}

inBuf := bufio.NewReader(cmd.InOrStdin())
name := args[0]

key, err := clientCtx.Keyring.Key(name)
if err != nil {
return errors.Wrapf(err, "failed to fetch '%s' from the keyring", name)
}

moniker := config.Moniker
if m, _ := cmd.Flags().GetString(stakingcli.FlagMoniker); m != "" {
moniker = m
}

createValCfg, err := stakingcli.PrepareConfigForTxCreateValidator(cmd.Flags(), moniker, nodeID, appGenesis.ChainID, valPubKey)
if err != nil {
return errors.Wrap(err, "error creating configuration to create validator msg")
}

amount := args[1]
coins, err := sdk.ParseCoinsNormalized(amount)
if err != nil {
return errors.Wrap(err, "failed to parse coins")
}

addr, err := key.GetAddress()
if err != nil {
return err
}

err = genutil.ValidateAccountInGenesis(genesisState, genBalIterator, addr, coins, cdc)
if err != nil {
return errors.Wrap(err, "failed to validate account in genesis")
}

txFactory, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
if err != nil {
return err
}

fromAddr, err := key.GetAddress()
if err != nil {
return err
}

clientCtx = clientCtx.WithInput(inBuf).WithFromAddress(fromAddr)

createValCfg.Amount = amount

txBldr, msg, err := stakingcli.BuildCreateValidatorMsg(clientCtx, createValCfg, txFactory, true, valAddressCodec)
if err != nil {
return errors.Wrap(err, "failed to build create-validator message")
}

if m, ok := msg.(*stakingtypes.MsgCreateValidator); ok {
m.DelegatorAddress = fromAddr.String()
msg = m
}

if key.GetType() == keyring.TypeOffline || key.GetType() == keyring.TypeMulti {
cmd.PrintErrln("Offline key passed in. Use `tx sign` command to sign.")
return txBldr.PrintUnsignedTx(clientCtx, msg)
}

w := bytes.NewBuffer([]byte{})
clientCtx = clientCtx.WithOutput(w)

if m, ok := msg.(sdk.HasValidateBasic); ok {
if err := m.ValidateBasic(); err != nil {
return err
}
}

if err = txBldr.PrintUnsignedTx(clientCtx, msg); err != nil {
return errors.Wrap(err, "failed to print unsigned std tx")
}

stdTx, err := readUnsignedGenTxFile(clientCtx, w)
if err != nil {
return errors.Wrap(err, "failed to read unsigned gen tx file")
}

txBuilder, err := clientCtx.TxConfig.WrapTxBuilder(stdTx)
if err != nil {
return fmt.Errorf("error creating tx builder: %w", err)
}

err = authclient.SignTx(txFactory, clientCtx, name, txBuilder, true, true)
if err != nil {
return errors.Wrap(err, "failed to sign std tx")
}

outputDocument, _ := cmd.Flags().GetString(flags.FlagOutputDocument)
if outputDocument == "" {
outputDocument, err = makeOutputFilepath(config.RootDir, nodeID)
if err != nil {
return errors.Wrap(err, "failed to create output file path")
}
}

if err := writeSignedGenTx(clientCtx, outputDocument, stdTx); err != nil {
return errors.Wrap(err, "failed to write signed gen tx")
}

cmd.PrintErrf("Genesis transaction written to %q`n", outputDocument)
return nil
},
}

cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
cmd.Flags().String(flags.FlagOutputDocument, "", "Write the genesis transaction JSON document to the given file instead of the default location")
cmd.Flags().AddFlagSet(fsCreateValidator)
flags.AddTxFlagsToCmd(cmd)
_ = cmd.Flags().MarkHidden(flags.FlagOutput)

return cmd
}

func makeOutputFilepath(rootDir, nodeID string) (string, error) {
writePath := filepath.Join(rootDir, "config", "gentx")
if err := os.MkdirAll(writePath, 0o700); err != nil {
return "", fmt.Errorf("could not create directory %q: %w", writePath, err)
}

return filepath.Join(writePath, fmt.Sprintf("gentx-%v.json", nodeID)), nil
}

func readUnsignedGenTxFile(clientCtx client.Context, r io.Reader) (sdk.Tx, error) {
bz, err := io.ReadAll(r)
if err != nil {
return nil, err
}

aTx, err := clientCtx.TxConfig.TxJSONDecoder()(bz)
if err != nil {
return nil, err
}

return aTx, err
}

func writeSignedGenTx(clientCtx client.Context, outputDocument string, tx sdk.Tx) error {
outputFile, err := os.OpenFile(outputDocument, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
if err != nil {
return err
}
defer outputFile.Close()

jsonBz, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
if err != nil {
return err
}

_, err = fmt.Fprintf(outputFile, "%s`n", jsonBz)
return err
}
