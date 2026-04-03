#/bin/sh

# KEY MANAGEMENT
KEYRING="test"

# Function updates the config based on a jq argument as a string
update_test_genesis () {
    # EX: update_test_genesis '.consensus_params["block"]["max_gas"]="100000000"'
    cat ~/.do/config/genesis.json | jq --arg DENOM "$2" "$1" > ~/.do/config/tmp_genesis.json && mv ~/.do/config/tmp_genesis.json ~/.do/config/genesis.json
}

# add keys, add balances
for i in $(seq 0 3); do
    key=$(jq ".keys[$i] | tostring" /keys.json )
    keyname=$(echo $key | jq -r 'fromjson | ."keyring-keyname"')
    mnemonic=$(echo $key | jq -r 'fromjson | .mnemonic')
    # Add new account
    echo $mnemonic | dochaind keys add $keyname --keyring-backend $KEYRING --recover --home ~/.do
    # Add initial balances
    dochaind add-genesis-account $keyname "1000000000000udo" --keyring-backend $KEYRING --home ~/.do
done

# Sign genesis transaction
dochaind gentx test "1000000udo" --keyring-backend $KEYRING --chain-id $CHAINID --home ~/.do

update_test_genesis '.app_state["gov"]["voting_params"]["voting_period"] = "50s"'
update_test_genesis '.app_state["mint"]["params"]["mint_denom"]=$DENOM' udo
update_test_genesis '.app_state["gov"]["deposit_params"]["min_deposit"]=[{"denom": $DENOM,"amount": "1000000"}]' udo
update_test_genesis '.app_state["crisis"]["constant_fee"]={"denom": $DENOM,"amount": "1000"}' udo
update_test_genesis '.app_state["staking"]["params"]["bond_denom"]=$DENOM' udo

# Collect genesis tx
dochaind collect-gentxs --home ~/.do

# Run this to ensure everything worked and that the genesis file is setup correctly
dochaind validate-genesis --home ~/.do




