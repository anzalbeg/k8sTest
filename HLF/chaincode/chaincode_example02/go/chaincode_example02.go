package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type PurchaseOrder struct {
	ObjectType         string             `json:"objectType"`
	PurchaseOrderID    string             `json:"purchaseOrderID"`
	Ref                string             `json:"ref"`
	OrderDate          time.Time          `json:"orderDate"`
	PurchaseOrderState PurchaseOrderState `json:"purchaseOrderState"`
}

type PurchaseOrderState int

const (
	ordered PurchaseOrderState = iota
	pendingDelivery
	received
	cancelled
)

type Shipment struct {
	ObjectType            string             `json:"objectType"`
	ShipmentID            string             `json:"shipmentID"`
	PurchaseOrder         PurchaseOrder      `json:"purchaseOrder"`
	Customer              Customer           `json:"customerID"`
	Carrier               Carrier            `json:"carrier"`
	Location              Location           `json:"location"`
	ExpectedDepartureDate time.Time          `json:"expectedDepartureDate"`
	ExpectedArrivedDate   time.Time          `json:"expectedArrivedDate"`
	RealDepartureDate     time.Time          `json:"realDepartureDate"`
	RealArrivedDate       time.Time          `json:"realArrivedDate"`
	ShipmentOrderState    ShipmentOrderState `json:"shipmentOrderState"`
	Dispute               bool               `json:"dispute"`
	ReasonDispute         string             `json:"reasonDispute"`
}

type ShipmentOrderState int

const (
	waiting ShipmentOrderState = iota
	loading
	loaded
	inTransit
	deliveredComplete
	deliveredIncomplete
)

type Customer struct {
	CustomerID string `json:"customerID"`
}

type Carrier struct {
	CarrierID string `json:"carrierID"`
}

type Location struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
	Address   string `json:"address"`
	Dock      string `json:"dock"`
}

// String returns the name of the state
func (n PurchaseOrderState) String() string {
	names := [...]string{"ordered", "pendingDelivery", "received", "cancelled"}

	// prevent panicking in case of Weekday is out-of-range
	if n < ordered || n > cancelled {
		return "Unknown"
	}

	return names[n]
}

func (n ShipmentOrderState) String() string {
	names := [...]string{"waiting", "loading", "loaded", "inTransit,", "deliveredComplete", "deliveredIncomplete"}

	// prevent panicking in case of Weekday is out-of-range
	if n < waiting || n > deliveredIncomplete {
		return "Unknown"
	}

	return names[n]
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "createShipment" { //create a new shipment
		return t.createShipment(stub, args)
	} else if function == "readShipmentData" { //read a shipment
		return t.readShipmentData(stub, args)
	} else if function == "getHistoryForShipment" { //get history of values for a shipment
		return t.getHistoryForShipment(stub, args)
	} else if function == "getShipmentByRange" { //get containers based on range query
		return t.getShipmentByRange(stub, args)
	} else if function == "transferShipment" { //change owner of a specific shipment
		return t.transferShipment(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initContainer - create a new shipment, store into chaincode state
// ============================================================
func (t *SimpleChaincode) createShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error

	fmt.Println(len(args))
	// ==== Input sanitation ====
	fmt.Println("- start init shipment")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}

	shipmentID := args[0]
	payload := args[1]

	fmt.Println("shipmentID--" + shipmentID)
	fmt.Println("Payload--" + payload)

	// ==== Check if shipmentID already exists ====
	shipmentAsBytes, err := stub.GetState(shipmentID)
	if err != nil {
		return shim.Error("Failed to get shipmentID: " + err.Error())
	} else if shipmentAsBytes != nil {
		fmt.Println("This shipmentID already exists: " + shipmentID)
		return shim.Error("This shipmentID already exists: " + shipmentID)
	}
	// ==== Create shipment object and shipment to JSON ====
	var shipmentVariable Shipment
	if err := json.Unmarshal([]byte(payload), &shipmentVariable); err != nil {
		log.Fatal(err)
	}
	purchaseOrderID := shipmentVariable.PurchaseOrder.PurchaseOrderID
	_tempPurchase := PurchaseOrder{
		ObjectType:         "purchaseOrder",
		PurchaseOrderID:    shipmentVariable.PurchaseOrder.PurchaseOrderID,
		Ref:                shipmentVariable.PurchaseOrder.Ref,
		OrderDate:          shipmentVariable.PurchaseOrder.OrderDate,
		PurchaseOrderState: shipmentVariable.PurchaseOrder.PurchaseOrderState,
	}

	_tempPurchaseJsonAsBytes, err := json.Marshal(_tempPurchase)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println(string(_tempPurchaseJsonAsBytes))
	_tempShipmentJsonAsBytes, err1 := json.Marshal(payload)
	// === Save shipment to state ===
	err = stub.PutState(purchaseOrderID, []byte(_tempPurchaseJsonAsBytes))
	if err != nil {
		return shim.Error(err.Error())
	}
	err1 = stub.PutState(shipmentID, []byte(_tempShipmentJsonAsBytes))
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	fmt.Println(" Succussfully created shipment details")
	return shim.Success(nil)
}

// ===============================================
// readShipmentData - read a shipment from chaincode state
// ===============================================
func (t *SimpleChaincode) readShipmentData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var shipmentID, shipmentDataJsonResp string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting name of the shipment to query")
	}

	shipmentID = args[0]
	shipmentDataAsbytes, err := stub.GetState(shipmentID) //get the shipment from chaincode state
	if err != nil {
		shipmentDataJsonResp = "{\"Error\":\"Failed to get state for " + shipmentID + "\"}"
		return shim.Error(shipmentDataJsonResp)
	} else if shipmentDataAsbytes == nil {
		shipmentDataJsonResp = "{\"Error\":\"shipment does not exist: " + shipmentID + "\"}"
		return shim.Error(shipmentDataJsonResp)
	}
	return shim.Success(shipmentDataAsbytes)
}

// ==================================================
// deleteShipment - remove a shipment key/value pair from state
// ==================================================

// // ===========================================================
// // transfer a shipment by setting a new owner name on the shipment
// // ===========================================================
func (t *SimpleChaincode) transferShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	//var shipmentStatus string
	shipmentID := args[0]
	payload := args[1]
	fmt.Println("- start updation ", shipmentID, payload)

	shipmentAsBytes, err := stub.GetState(shipmentID)
	if err != nil {
		return shim.Error("Failed to get shipmentID:" + err.Error())
	} else if shipmentAsBytes == nil {
		return shim.Error("shipmentID does not exist")
	}

	var _tempShipment Shipment
	//surgicalkitToTransfer := surgicalkit{}
	err1 := json.Unmarshal(shipmentAsBytes, &_tempShipment) //unmarshal it aka JSON.parse()
	if err1 != nil {
		return shim.Error(err1.Error())
	}
	//_tempShipment.shipmentID = shipmentID //change the owner
	//_tempShipment.Payload 	 = Payload
	//var j = []byte(_tempShipment)
	// a map shipment to decode the JSON structure into
	//c := make(map[string]interface{})

	// unmarschal JSON
	//  json.Unmarshal(j, &c)
	// for s, v := range c {
	// 	switch vv :=v.(type) {
	// 		case string :
	// 					if(s=="st"){
	// 						fmt.Println("status field found in payload--"+vv)
	// 						shipmentStatus = vv
	// 					}
	// 	}
	// }
	//fmt.Println("shipmentStatus------"+shipmentStatus)

	// if shipmentStatus == "tampered" {
	// 	surgicalkitToTransfer.Compliant = false
	// }else{
	// 	surgicalkitToTransfer.Compliant = true
	// }
	purchaseOrderID := _tempShipment.PurchaseOrder.PurchaseOrderID
	_tempPurchase := PurchaseOrder{
		ObjectType:         "purchaseOrder",
		PurchaseOrderID:    _tempShipment.PurchaseOrder.PurchaseOrderID,
		Ref:                _tempShipment.PurchaseOrder.Ref,
		OrderDate:          _tempShipment.PurchaseOrder.OrderDate,
		PurchaseOrderState: _tempShipment.PurchaseOrder.PurchaseOrderState,
	}
	_tempPurchaseJsonAsBytes, err2 := json.Marshal(_tempPurchase)
	if err2 != nil {
		return shim.Error(err2.Error())
	}
	fmt.Println("transfer ower-purchaseOrder-%v+", string(_tempPurchaseJsonAsBytes))
	// === Save shipment to state ===
	err3 := stub.PutState(purchaseOrderID, []byte(_tempPurchaseJsonAsBytes))
	if err3 != nil {
		return shim.Error(err3.Error())
	}
	_tempShipmentJSONasBytes, _ := json.Marshal(_tempShipment)
	err4 := stub.PutState(shipmentID, _tempShipmentJSONasBytes) //rewrite the shipment
	if err4 != nil {
		return shim.Error(err4.Error())
	}
	fmt.Println("transfer ower-shipement-%v+", string(_tempShipmentJSONasBytes))
	fmt.Println("- end shipement transfer(success)")
	return shim.Success(nil)
}

func (t *SimpleChaincode) getShipmentByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	startKey := args[0]
	endKey := args[1]

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getShipmentByRange queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (t *SimpleChaincode) getHistoryForShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	shipmentID := args[0]

	fmt.Printf("- start getHistoryForShipment: %s\n", shipmentID)

	resultsIterator, err := stub.GetHistoryForKey(shipmentID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the surgicalkit
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxID\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a deleteShipment operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON shipment)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForShipment returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}
