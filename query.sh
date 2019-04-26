#!/bin/sh

# setting up the peer1 env variables inside cli environment
my_dir="$(dirname "$0")"
. "$my_dir/parse_yaml.sh"

eval $(parse_yaml fabric-artifacts/values.yaml "config_")
export GENESIS_BLOCK=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/channel-artifacts/orderer-channel.tx
export CLI_POD_ID=`kubectl get pod -n org1namespace | grep cli | cut -f1 -d' '`
export ORDERER_ADDR=$config_clusterIpRange".249.66:7050"
export ORG_DOMAIN="org1.example.com"
export CHAINCODE_PATH=github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/chaincode/chaincode_example02/go
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/crypto-config/peerOrganizations/$ORG_DOMAIN/users/Admin@$ORG_DOMAIN/msp
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_ADDRESS=$config_clusterIpRange".241.223:7051"
export CHANNEL_NAME="mychannel"
# invoke 

kubectl exec $CLI_POD_ID -n org1namespace -it -- bash -c "CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID && CORE_PEER_MSPCONFIGPATH=$CORE_PEER_MSPCONFIGPATH && CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS && peer chaincode query -C $CHANNEL_NAME -n supplychain -c '{\"Args\":[\"readShipmentData\",\"shipment01\"]}'"


#peer0org2
export GENESIS_BLOCK=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/channel-artifacts/orderer-channel.tx
export CLI_POD_ID=`kubectl get pod -n org1namespace | grep cli | cut -f1 -d' '`
export ORDERER_ADDR=$config_clusterIpRange".249.66:7050"
export ORG_DOMAIN="org2.example.com"
export CHAINCODE_PATH=github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/chaincode/chaincode_example02/go
export CORE_PEER_MSPCONFIGPATH="/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/crypto-config/peerOrganizations/$ORG_DOMAIN/users/Admin@$ORG_DOMAIN/msp"
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_ADDRESS=$config_clusterIpRange".245.74:7051"
export CHANNEL_NAME="mychannel"

kubectl exec $CLI_POD_ID -n org1namespace -it -- bash -c "CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID && CORE_PEER_MSPCONFIGPATH=$CORE_PEER_MSPCONFIGPATH && CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS && peer chaincode query -C $CHANNEL_NAME -n supplychain -c '{\"Args\":[\"readShipmentData\",\"shipment01\"]}'"
