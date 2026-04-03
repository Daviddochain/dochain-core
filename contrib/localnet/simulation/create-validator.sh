#!/bin/sh

# create a new validator node in local
# if /Users/thevinhnguyen/.dochain/config/priv_validator_key.json exists
# then remove it
if [ ! -f $NODE_HOME/config/priv_validator_key.json ]; then
    dochaind init moniker --chain-id $CHAIN_ID --home $NODE_HOME
fi

echo $CHAIN_ID

# create a new validator
dochaind keys add validator --keyring-backend $KEYRING_BACKEND --home $NODE_HOME

# fund the validator
dochaind tx bank send test0 $(dochaind keys show validator -a --keyring-backend $KEYRING_BACKEND --home $NODE_HOME) 50000000uluna --chain-id $CHAIN_ID --keyring-backend $KEYRING_BACKEND --home $NODE_HOME --node $(sh $SIMULATION_FOLDER/next_node.sh) --gas auto --gas-adjustment 2.3 --fees 20000000uluna -y

sleep 10

# create a validator for a node
dochaind tx staking create-validator --moniker test0 \
--from validator \
--amount="1000000uluna" \
--fees 20000000uluna \
--pubkey="$(dochaind tendermint show-validator --home $NODE_HOME)" \
--details="this is a validator" \
--commission-max-rate="0.10" \
--commission-max-change-rate="0.05" \
--commission-rate="0.05" \
--min-self-delegation 1 \
--chain-id $CHAIN_ID \
--keyring-backend $KEYRING_BACKEND \
--home $NODE_HOME \
--node $(sh $SIMULATION_FOLDER/next_node.sh) \
--gas auto \
--gas-adjustment 2.3 \
-y

sleep 10

# check if command `dochaind q staking validator $(dochaind keys show test0 -a --bech val --keyring-backend test)` success
dochaind q staking validator $(dochaind keys show test0 -a --bech val --keyring-backend test --home $NODE_HOME) >/dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "VALIDATOR CREATED SUCCESSFULLY"
else
    echo "FAILED TO CREATE VALIDATOR"
fi



