#!/bin/sh

# export VERSION=1.1.1
# export ARCH=$(echo "$(uname -s|tr '[:upper:]' '[:lower:]'|sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')

# BIN="$ARCH-$VERSION"
# #SLEEP_TIMEOUT=0

# #if [ ! -d "bin"]; then
#  echo "heelo bin"
#    curl https://nexus.hyperledger.org/content/repositories/releases/org/hyperledger/fabric/hyperledger-fabric/${BIN}/hyperledger-fabric-${BIN}.tar.gz | tar xz
# #fi

# ./bin/cryptogen generate --config=./crypto-config.yaml

export FABRIC_CFG_PATH=$PWD

OUTPUT_DIR=$PWD/channel-artifacts

mkdir -p $OUTPUT_DIR

./bin/configtxgen -profile TwoOrgsOrdererGenesis -outputBlock $OUTPUT_DIR/orderer-genesis.block
./bin/configtxgen -profile TwoOrgsChannel -outputCreateChannelTx $OUTPUT_DIR/orderer-channel.tx -channelID mychannel
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate $OUTPUT_DIR/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
./bin/configtxgen -profile TwoOrgsChannel -outputAnchorPeersUpdate $OUTPUT_DIR/Org2MSPanchors.tx -channelID mychannel -asOrg Org2MSP

echo "... 🙌🏿"
  