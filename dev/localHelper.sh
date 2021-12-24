#!/bin/bash


KEY="captain"
CHAINID="exchain-67"
MONIKER="oec"
CURDIR=`dirname $0`
HOME_SERVER=/Users/lvcong/go/src/github.com/okex/exchain/_cache_evm


cat /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="okt"' > /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json
cat /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="okt"' > /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json
cat /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="okt"' > /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json
cat /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="okt"' > /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json


sed -i "" 's/"enable_call": false/"enable_call": true/' /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json
sed -i "" 's/"enable_create": false/"enable_create": true/' /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json
sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm/config/genesis.json

# ex1h0j8x0v9hs4eq6ppgamemfyu4vuvp2sl0q9p3v
exchaind add-genesis-account $(exchaincli keys show 'captain'    -a) 100000000okt --home $HOME_SERVER
#ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9
exchaind add-genesis-account ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9 900000000okt --home $HOME_SERVER
set -o

