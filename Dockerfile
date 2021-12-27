# Simple usage with a mounted data directory:
# > docker build -t exchain .
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.exchaind:/root/.exchaind -v ~/.exchaincli:/root/.exchaincli exchain exchaind init mynode
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.exchaind:/root/.exchaind -v ~/.exchaincli:/root/.exchaincli exchain exchaind start
FROM okechain/enviroment:v0.0.1 AS build-env

# Install minimum necessary dependencies, remove packages
RUN apk add --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

# Set working directory for the build
WORKDIR /go/src/github.com/okex/exchain

# Add source files
COPY . .

# Build OKExChain
RUN GOPROXY=http://goproxy.cn make install

# Final image
FROM alpine:edge

ENV LOG_LEVEL=main:info,iavl:info,*:error,state:info
ENV HOME_SERVER=/root/
ENV CHAINID=exchain-67

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/exchaind /usr/bin/exchaind
COPY --from=build-env /go/bin/exchaincli /usr/bin/exchaincli

# Run exchaind by default, omit entrypoint to ease using container with exchaincli
CMD ["exchaind start --pruning=nothing --rpc.unsafe \
                     --local-rpc-port 26657 \
                     --log_level $LOG_LEVEL \
                     --consensus.timeout_commit 600ms \
                     --enable-preruntx \
                     --iavl-enable-async-commit \
                     --iavl-enable-gid \
                     --iavl-commit-interval-height 10 \
                     --iavl-output-modules evm=1,acc=1 \
                     --trace --home $HOME_SERVER --chain-id $CHAINID \
                     --elapsed Round=1,CommitRound=1,Produce=1 \
                     --rest.laddr 'tcp://localhost:8545'"]
