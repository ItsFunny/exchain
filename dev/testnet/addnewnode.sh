#!/usr/bin/env bash

source exchain.profile

set -e
set -o errexit
set -a
set -m

set -x # activate debugging


while getopts "i:n:p:r:s:b:" opt; do
  case $opt in
    i)
      echo "IP=$OPTARG"
      IP=$OPTARG
      ;;
    n)
      echo "INPUT_INDEX=$OPTARG"
      INPUT_INDEX=$OPTARG
      ;;
    p)
      echo "INPUT_P2PPORT=$OPTARG"
      INPUT_P2PPORT=$OPTARG
      ;;
    r)
      echo "INPUT_RPCPORT=$OPTARG"
      INPUT_RPCPORT=$OPTARG
      ;;
    s)
      echo "INPUT_SEEDNODE=$OPTARG"
      INPUT_SEEDNODE=$OPTARG
      ;;
    b)
      echo "BIN_NAME=$OPTARG"
      BIN_NAME=$OPTARG
      ;;
    \?)
      echo "Invalid option: -$OPTARG"
      ;;
  esac
done

function usage {
    echo "Invalid index!"
    echo "Use '-n' to specify the node id. Less then 99."
}

if [ -z "${INPUT_INDEX}" ]; then
    usage
    exit
fi

NAME=node${INPUT_INDEX}
let p2p_port=${BASE_PORT_PREFIX}+${INPUT_INDEX}*100+${P2P_PORT_SUFFIX}
let rpc_port=${BASE_PORT_PREFIX}+${INPUT_INDEX}*100+${RPC_PORT_SUFFIX}

# overwrite default ones
if [ ! -z ${INPUT_RPCPORT} ]; then
    rpc_port=${INPUT_RPCPORT}
fi

if [ ! -z ${INPUT_P2PPORT} ]; then
    p2p_port=${INPUT_P2PPORT}
fi

if [ -z ${IP} ]; then
    IP="127.0.0.1"
fi

if [ -d ${OKCHAIN_NET_CACHE}/node0/exchaind ]; then
    seed_addr=$(${BIN_NAME} tendermint show-node-id --home ${OKCHAIN_NET_CACHE}/node0/exchaind)@${IP}:${seedp2pport}
fi

if [ ! -z ${INPUT_SEEDNODE} ]; then
    seed_addr=${INPUT_SEEDNODE}
fi


init() {
    if [ ${INPUT_INDEX} -gt 99 ]; then
        usage
        exit
    fi

    if [ "${INPUT_INDEX}" -lt 1 ]; then
        usage
        exit
    fi

    if [ -d ${OKCHAIN_NET_CACHE}/${NAME} ]; then
        echo "Invalid index!"
        echo "<${OKCHAIN_NET_CACHE}/${NAME}> already exists. Use '-n' to try another index."
        echo "For example: ./addnewnode.sh -n 9 -s ${seed_addr}"
        exit
    fi

    if [ -z ${seed_addr} ]; then
        echo "Invalid seed node!"
        echo "Use '-s' to specify the seed node."
        echo "For example: ./addnewnode.sh -n 6 -s ${seed_addr}"
        exit
    fi

    ${BIN_NAME} init ${NAME} -o --chain-id ${CHAIN_ID} --home ${OKCHAIN_NET_CACHE}/${NAME}/exchaind
}


start() {
    echo "init new node..."
    init
    echo "init new node done"


    echo "copy the genesis file..."
    rm ${OKCHAIN_NET_CACHE}/${NAME}/exchaind/config/genesis.json
    cp ${OKCHAIN_NET_CACHE}/node0/exchaind/config/genesis.json ${OKCHAIN_NET_CACHE}/${NAME}/exchaind/config/
    echo "copy the genesis file done"

    echo "start new node..."
    p2pport=$1
    rpcport=$2
    seednode=$3

#     echo "${BIN_NAME} --home ${OKCHAIN_NET_CACHE}/${NAME}/exchaind  start --p2p.laddr tcp://${IP}:${p2pport} --p2p.seeds ${seednode} --rpc.laddr tcp://${IP}:${rpcport}"

#    LOG_LEVEL=main:info,*:error
#    LOG_LEVEL=main:info,*:error,consensus:info,state:info
    LOG_LEVEL=main:info,*:error,state:info

    ${BIN_NAME} start \
    --chain-id ${CHAIN_ID} \
    --home ${OKCHAIN_NET_CACHE}/${NAME}/exchaind \
    --p2p.laddr tcp://${IP}:${p2pport} \
    --p2p.seeds ${seednode} \
    --log_level ${LOG_LEVEL} \
    --download-delta \
    --p2p.addr_book_strict=false \
    --rpc.laddr tcp://${IP}:${rpcport} > ${OKCHAIN_NET_CACHE}/rpc${INPUT_INDEX}.log 2>&1 &

#     echo "start new node done"
#     --download-delta \
#     --enable-proactively-runtx \

}


start ${p2p_port} ${rpc_port} ${seed_addr}
