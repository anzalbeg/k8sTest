#!/bin/sh

# setting up the peer1 env variables inside cli environment

my_dir="$(dirname "$0")"
. "$my_dir/parse_yaml.sh"

eval $(parse_yaml fabric-artifacts/values.yaml "config_")

# setting up the peer1 env variables inside cli environment
export GENESIS_BLOCK=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/channel-artifacts/orderer-channel.tx
export CLI_POD_ID=`kubectl get pod | grep cli | cut -f1 -d' '`
export ORDERER_ADDR="dev-orderer-g746q:7050"
export ORG_DOMAIN="org1.example.com"
export CHAINCODE_PATH=github.com/hyperledger/fabric/peer/crypto/crypto-config/opensource.com/HLF/chaincode/chaincode_example02/go
export CORE_PEER_MSPCONFIGPATH=/etc/crypto-config/opensource.com/HLF/crypto-config/peerOrganizations/$ORG_DOMAIN/users/Admin@$ORG_DOMAIN/msp
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_ADDRESS="dev-peer0-xyz:7051"
export CHANNEL_NAME="mychannel"
# invoke 

set -x
kubectl exec $CLI_POD_ID -it -- bash -c "CORE_PEER_LOCALMSPID=$CORE_PEER_LOCALMSPID && CORE_PEER_ADDRESS=$CORE_PEER_ADDRESS && peer chaincode invoke -o $ORDERER_ADDR -C $CHANNEL_NAME -n supplychain -v 1.0 -c '{\"Args\":[\"createShipment\",\"shipment01\",\"{\\\"objectType\\\": \\\"shipment\\\",\\\"shipmentID\\\": \\\"shipment01\\\",\\\"purchaseOrder\\\": {\\\"purchaseOrderID\\\": \\\"purchase01\\\", \\\"ref\\\": \\\"987667eye56728yx87q80j\\\", \\\"shipmentOrderedState\\\": \\\"deliverywaiting\\\", \\\"orderDate\\\": \\\"2018-06-05T17:00:00Z\\\" }, \\\"customer\\\": { \\\"customerID\\\": \\\"customerID01\\\" }, \\\"carrier\\\": { \\\"carrierID\\\": \\\"3rdPartyLogistic\\\" }, \\\"location\\\": { \\\"latitude\\\": \\\"48.8566\\\", \\\"longitude\\\": \\\"2.3522\\\", \\\"address\\\": \\\"paris\\\", \\\"dock\\\": \\\"ns\\\" }, \\\"expectedDepartureDate\\\": \\\"2018-06-05T17:00:00Z\\\", \\\"expectedArrivedDate\\\": \\\"2018-06-05T17:00:00Z\\\", \\\"realDepartureDate\\\": \\\"2018-06-05T17:00:00Z\\\", \\\"realArrivedDate\\\": \\\"2018-06-05T17:00:00Z\\\", \\\"shipmentOrderedState\\\": \\\"loaded\\\", \\\"dispute\\\": false, \\\"reasonDispute\\\": \\\"na\\\" }\"]}'"
set +x
