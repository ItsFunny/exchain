#!/bin/bash

PREFIX_PATH=${GOPATH}

KEY="captain"
CHAINID="exchain-67"
MONIKER="oec"
CURDIR=`dirname $0`
HOME_SERVER=${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm

function  error()  {
    if [ $? -ne 0 ];then
        echo ${1}
        exit 0
    fi
}


rm -rf ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm

set -o
# 1. init
exchaincli config chain-id exchain-67
error 1
# 2. output json
exchaincli config output json
error 2
# 3. config indent true
exchaincli config indent true
error 3
# 4. config trust-node true
exchaincli config trust-node true
error 4
# 5. config keyring-backend test
exchaincli config keyring-backend test
error 5
# 6. keys add --recover captain -m "puzzle glide follow cruel say burst deliver wild tragic galaxy lumber offer" -y
exchaincli keys add --recover captain -m "puzzle glide follow cruel say burst deliver wild tragic galaxy lumber offer" -y
error 6
# 7. keys add --recover admin16 -m "palace cube bitter light woman side pave cereal donor bronze twice work" -y
exchaincli keys add --recover admin16 -m "palace cube bitter light woman side pave cereal donor bronze twice work" -y
error 7
# 8. init oec --chain-id exchain-67 --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
exchaind init oec --chain-id exchain-67 --home ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm
error 8
# 9. add-genesis-account ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9 100000000okt --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
exchaind add-genesis-account ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9 100000000okt --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
error 9
# 10. add genesis account 2
exchaind  add-genesis-account ex1h0j8x0v9hs4eq6ppgamemfyu4vuvp2sl0q9p3v 100000000okt --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
error 10
# 11. gentx --name captain --keyring-backend test --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
exchaind gentx --name captain --keyring-backend test --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
error 11
# 12. exchaind testnet --v 4 -o cache -l --chain-id exchainevm-8 --starting-ip-address 127.0.0.1 --base-port 10056 --keyring-backend test --mnemonic=false
exchaind testnet --v 4 -o cache -l --chain-id exchainevm-8 --starting-ip-address 127.0.0.1 --base-port 10056 --keyring-backend test --mnemonic=false
error 12
# 13. validate-genesis --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
exchaind validate-genesis --home /Users/lvcong/go/src/github.com/okex/exchain/_cache_evm
error 13
# 14. config keyring-backend test
exchaincli config keyring-backend test
error

#
#cat ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="okt"' > ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#cat ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="okt"' > ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#cat ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="okt"' > ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#cat ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="okt"' > ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json && mv ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/tmp_genesis.json ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#
#
#sed -i "" 's/"enable_call": false/"enable_call": true/' ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#sed -i "" 's/"enable_create": false/"enable_create": true/' ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#sed -i "" 's/"enable_contract_blocked_list": false/"enable_contract_blocked_list": true/' ${PREFIX_PATH}/src/github.com/okex/exchain/_cache_evm/config/genesis.json
#
## ex1h0j8x0v9hs4eq6ppgamemfyu4vuvp2sl0q9p3v
#exchaind add-genesis-account $(exchaincli keys show 'captain'    -a) 100000000okt --home $HOME_SERVER
##ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9
#exchaind add-genesis-account ex1s0vrf96rrsknl64jj65lhf89ltwj7lksr7m3r9 900000000okt --home $HOME_SERVER
#
#
#
set +o
#
