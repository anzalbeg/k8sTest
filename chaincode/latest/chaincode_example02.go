package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SupplyChainChaincode example simple Chaincode implementation
type SupplyChainChaincode struct {
}

type Organization struct {
	OrganizationID string `json:"organizationID"`
	Tx             string `json:"tx"`
	Name           string `json:"name"`
}

type ParticipantUser struct {
	ParticipantID string       `json:"participantID"`
	Name          string       `json:"name"`
	Organization  Organization `json:"organization"`
	Role          Role         `json:"role"`
}

type Role int

const (
	customer Role = iota
	seller
	driver
	LogisticArgument
	LogisticManager
)

type Logistics struct {
	LogisticsID   string             `json:"logisticsID"`
	Contains      []LogisticsUnit    `json:"contains"`
	ContainedBY   []LogisticsUnit    `json:"containedBY"`
	State         PurchaseOrderState `json:"state"`
	Type          string             `json:"type"`
	Assignee      ParticipantUser    `json:"assignee"`
	Owner         ParticipantUser    `json:"owner"`
	CounterSignee ParticipantUser    `json:"counterSignee"`
	Location      Location           `json:"location"`
	Size          float64            `json:"size"`
	Weight        float64            `json:"weight"`
	Price         float64            `json:"price"`
}

type PurchaseOrder struct {
	PurchaseOrderID string             `json:"purchaseOrderID"`
	Seller          ParticipantUser    `json:"seller"`
	Buyer           ParticipantUser    `json:"buyer"`
	ExpectedDelDate time.Time          `json:"expectedDelDate"`
	ShipTO          Location           `json:"shipTO"`
	Amount          float64            `json:"amount"`
	Product         []LogisticsUnit    `json:"product"`
	State           PurchaseOrderState `json:"state"`
}

type LogisticsUnit struct {
	LogisticsUnitID string             `json:"logisticsUnitID"`
	ParentID        string             `json:"parentID"`
	ShipmentID      string             `json:"shipmentID"`
	LogisticsState  PurchaseOrderState `json:"state"`
	Type            string             `json:"type"`
	Assignee        ParticipantUser    `json:"assignee"`
	CounterSignee   ParticipantUser    `json:"counterSignee"`
	Location        Location           `json:"location"`
	Size            float64            `json:"size"`
	Weight          float64            `json:"weight"`
	PurchaseOrderID string             `json:"purchaseOrderID"`
	Quantity        int                `json:"quantity"`
	Price           float64            `json:"price"`
	DisputeReason   string             `json:"disputeReason"`
	DisputeComment  string             `json:"disputeComment"`
}

type Location struct {
	LocationID     string `json:"locationID"`
	Street         string `json:"street"`
	DockLineNumber string `json:"dockLineNumber"`
	PostalCode     string `json:"postalCode"`
	City           string `json:"city"`
	Country        string `json:"country"`
}

type PurchaseOrderState int

const (
	AwaitingValidation PurchaseOrderState = iota
	Validated
	Prepared
	Shipped
	Delivered
	Rejected
	Paid
	Pending
	AwaitingPayment
)

type Shipment struct {
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

// String returns the name of the state
func (n PurchaseOrderState) String() string {
	names := [...]string{"AwaitingValidation", "Validated", "Prepared", "Shipped", "Delivered", "Rejected", "Paid", "Pending", "AwaitingPayment"}
	// prevent panicking in case of Weekday is out-of-range
	if n < AwaitingValidation || n > AwaitingPayment {
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
	err := shim.Start(new(SupplyChainChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SupplyChainChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SupplyChainChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
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
	} else if function == "createOrganization" { //create new organization
		return t.createOrganization(stub, args)
	} else if function == "getOrganizationbyID" { //create new organization
		return t.getOrganizationbyID(stub, args)
	}
	// else if function == "createParticipantUser" { // create a participant within the organization
	// 	return t.createUser(stub, args)
	// } else if function == "assignSecurityRole" { // assign or update the rol of participant
	// 	return t.assignSecurityRole(stub, args)
	// } else if function == "createLogisticUnit" { // create logisticunit
	// 	return t.createLogisticUnit(stub, args)
	// } else if function == "packageLogistic" { // packaging the logistic unit
	// 	return t.packageLogistic(stub, args)
	// }

	fmt.Println("invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initContainer - create a new shipment, store into chaincode state
// ============================================================
// func newUUID() (string, error) {
// 	uuid := make([]byte, 16)
// 	n, err := io.ReadFull(rand.Reader, uuid)
// 	if n != len(uuid) || err != nil {
// 		return "", err
// 	}
// 	// variant bits; see section 4.1.1
// 	uuid[8] = uuid[8]&^0xc0 | 0x80
// 	// version 4 (pseudo-random); see section 4.1.3
// 	uuid[6] = uuid[6]&^0xf0 | 0x40
// 	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
// }

func (t *SupplyChainChaincode) getOrganizationbyID(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp, errResp string
	fmt.Println("start getOrganizationbyID", args[0])
	var err error
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 1\" \" as an argument\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
	_organizationID := args[0]

	_tempJSON := Organization{}

	jsonResp = ""
	valueAsBytes, err := stub.GetState(_organizationID)
	if err != nil {
		errResp = "{\"Error\":\"Failed to get state for " + _organizationID + "\"}"
		return shim.Error(errResp)
	}
	json.Unmarshal(valueAsBytes, &_tempJSON)

	if _tempJSON.OrganizationID == _organizationID {
		jsonResp = string(valueAsBytes[:])
	} else {
		fmt.Println(_organizationID + " not found")
		errMsg := "{ \"Organization\" : \"" + _organizationID + "\",\"message\" : \"" + "Organization Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
	}
	fmt.Println("jsonResp : " + jsonResp)
	fmt.Println("end getOrganizationbyID")
	return shim.Success([]byte(jsonResp))
	//send it onward
}

func (t *SupplyChainChaincode) createOrganization(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var tosend, organizationID string
	var uuid [16]byte
	fmt.Println("- start creating organization")

	_payloadBytes, _ := json.Marshal(args[0])

	uuid = md5.Sum(_payloadBytes)

	organizationID = string(uuid[:])

	fmt.Printf("uuid value is : %s\n", uuid)
	fmt.Println("- start creating transaction")
	err = stub.PutState(organizationID, []byte(args[0]))
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("- transaction created")

	tosend = "{  \"message\" : \"Organization created successfully on legder with organizationID \" : \"" + organizationID + "\", \"code\" : \"200\"}"
	//}else if count == len(existPolicyArray){
	//tosend = "{ \"message\" : \"All of these Polices already exists in legder.\", \"Policies failed to stored on ledger with policyIds \" : \"" + strings.Join(existPolicyArray, ",") + "\", \"code\" : \"503\"}"
	//}else{
	//tosend = "{ \"message\" : \"Polices stored successfully on legder except few policies which already exists in ledger. Please find below details for more information.\", \"Policies successfully stored on ledger with policyIds \" : \"" + strings.Join(policyIdArray, ",") + "\", \"Policies failed to stored on ledger with policyIds \" : \"" + strings.Join(existPolicyArray, ",") + "\", \"Properties successfully created/updated with propertyIds\" : \"" + strings.Join(propertyIdArray, ",") + "\", \"code\" : \"503\"}"
	//}

	fmt.Println("event message: " + tosend)
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *SupplyChainChaincode) createShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
	_tempPurchase := &PurchaseOrder{PurchaseOrderID: shipmentVariable.PurchaseOrder.PurchaseOrderID, State: shipmentVariable.PurchaseOrder.State}

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
func (t *SupplyChainChaincode) readShipmentData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
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
func (t *SupplyChainChaincode) transferShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
	_tempPurchase := &PurchaseOrder{PurchaseOrderID: _tempShipment.PurchaseOrder.PurchaseOrderID, State: _tempShipment.PurchaseOrder.State}
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

func (t *SupplyChainChaincode) getShipmentByRange(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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

func (t *SupplyChainChaincode) getHistoryForShipment(stub shim.ChaincodeStubInterface, args []string) pb.Response {

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
