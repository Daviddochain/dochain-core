#!/bin/sh

export SIMULATION_FOLDER=$(dirname $(realpath "$0"))
export TESTNET_FOLDER=$(echo $SIMULATION_FOLDER | sed 's|\(.*core\).*|\1|')/build
export NODE_HOME=${NODE_HOME:-$TESTNET_FOLDER/.do}
export KEYRING_BACKEND=test
export CHAIN_ID=${CHAIN_ID:-localdo}

echo $CHAIN_ID

if [ ! -d "$NODE_HOME" ]; then
    dochaind init moniker --chain-id $CHAIN_ID --home $NODE_HOME
fi

# initialize keys
for i in $(seq 0 3); do
    # delete all keys
    keys=$(dochaind keys list -n --keyring-backend $KEYRING_BACKEND --home $NODE_HOME)
    for key in $keys; do
        echo "y" | dochaind keys delete $key --keyring-backend $KEYRING_BACKEND --home $NODE_HOME
    done

    key=$(jq ".keys[$i] | tostring" $SIMULATION_FOLDER/network/$CHAIN_ID/keys.json )
    keyname=$(echo $key | jq -r 'fromjson | ."keyring-keyname"')
    mnemonic=$(echo $key | jq -r 'fromjson | .mnemonic')
    # Add new account
    echo $mnemonic | dochaind keys add $keyname --keyring-backend $KEYRING_BACKEND --home $NODE_HOME --recover
done

# if chain-id is localdo
if [ "$CHAIN_ID" = "localdo" ]; then
    # copy genesis.json to $NODE_HOME
    cp $TESTNET_FOLDER/node0/dochaind/config/genesis.json $NODE_HOME/config/genesis.json

    # add validator addresses
    # delete all keys
    keys=$(dochaind keys list -n --keyring-backend $KEYRING_BACKEND --home $NODE_HOME)
    for key in $keys; do
        echo "y" | dochaind keys delete $key --keyring-backend $KEYRING_BACKEND --home $NODE_HOME
    done

    for folder in "${TESTNET_FOLDER}"/node*/
    do
        position=$(basename $folder)
        position=${position:4}
        mnemonic=$(jq -r '.secret' ${folder}dochaind/key_seed.json)

        # Add new account
        echo $mnemonic | dochaind keys add test$position --keyring-backend $KEYRING_BACKEND --home $NODE_HOME --recover
    done
fi

# tx_send
sh $SIMULATION_FOLDER/tx_send.sh

echo "DONE TX SEND SIMULATION (1/5)"

# create-validator
sh $SIMULATION_FOLDER/create-validator.sh

echo "DONE CREATE VALIDATOR SIMULATION (2/5)"

# delegate
sh $SIMULATION_FOLDER/delegate.sh

echo "DONE DELEGATION SIMULATION (3/5)"

# contracts
sh $SIMULATION_FOLDER/contract.sh

echo "DONE CONTRACT SIMULATION (4/5)"

#governance
sh $SIMULATION_FOLDER/gov.sh

echo "DONE GOV SIMULATION (5/5)"






