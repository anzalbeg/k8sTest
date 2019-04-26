package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

type ManageProperties struct {
}

type PolicyImage struct {
	DocumentIdHash string `json:"documentIdHash"`
	FileName       string `json:"fileName,omitempty"`
	Format         string `json:"format,omitempty"`
	FileSizeBytes  int    `json:"fileSizeBytes,omitempty"`
}

type Policy struct {
	ClaimFlag               *bool       `json:"claimFlag,omitempty"`
	PolicyImage             PolicyImage `json:"policyImage"`
	EffectiveDate           time.Time   `json:"effectiveDate,omitempty"`
	EffectiveDateText       string      `json:"effectiveDateText,omitempty"`
	Endorsements            string      `json:"endorsements,omitempty"`
	Estate                  string      `json:"estate,omitempty"`
	Exceptions              string      `json:"exceptions,omitempty"`
	HOA                     string      `json:"hOA,omitempty"`
	InsuredMortgage         string      `json:"insuredMortgage,omitempty"`
	InsuredNames            string      `json:"insuredNames,omitempty"`
	IssueDate               time.Time   `json:"issueDate,omitempty"`
	IssuingCompanyName      string      `json:"issuingCompanyName,omitempty"`
	LegalDescription        string      `json:"legalDescription,omitempty"`
	LiabilityAmount         string      `json:"liabilityAmount,omitempty"`
	ObjectType              string      `json:"docType"` //docType is used to distinguish the various types of objects in state database
	OrganizationName        string      `json:"organizationName,omitempty"`
	OrganizationDataPrivate string      `json:"organizationDataPrivate,omitempty"`
	PropertyId              string      `json:"propertyId"`
	PolicyId                string      `json:"policyId"`
	PolicyKingdom           string      `json:"policyKingdom,omitempty"`
	PolicyName              string      `json:"policyName,omitempty"`
	PolicyNumber            string      `json:"policyNumber,omitempty"`
	PolicyType              string      `json:"policyType,omitempty"`
	SearchDepth             string      `json:"searchDepth,omitempty"`
	Subordinate             string      `json:"subordinate,omitempty"`
	TitleVesting            string      `json:"titleVesting,omitempty"`
	OrganizationId          string      `json:"organizationId,omitempty"`
}

type Property struct {
	AddressLine1      string `json:"addressLine1,omitempty"`
	AddressLine2      string `json:"addressLine2,omitempty"`
	City              string `json:"city,omitempty"`
	County            string `json:"county,omitempty"`
	ObjectType        string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	OrganizationName  string `json:"organizationName,omitempty"`
	PropertyId        string `json:"propertyId"`
	RelatedPropertyId string `json:"relatedPropertyId,omitempty"`
	State             string `json:"state,omitempty"`
	TaxID             string `json:"taxID,omitempty"`
	Zip               string `json:"zip,omitempty"`
	OrganizationId    string `json:"organizationId,omitempty"`
}

type IntermediateProperty struct {
	AddressLine1      string   `json:"addressLine1,omitempty"`
	AddressLine2      string   `json:"addressLine2,omitempty"`
	City              string   `json:"city,omitempty"`
	County            string   `json:"county,omitempty"`
	OrganizationName  string   `json:"organizationName,omitempty"`
	Policies          []Policy `json:"policies"`
	PropertyId        string   `json:"propertyId"`
	RelatedPropertyId string   `json:"relatedPropertyId,omitempty"`
	State             string   `json:"state,omitempty"`
	TaxID             string   `json:"taxID,omitempty"`
	Zip               string   `json:"zip,omitempty"`
	OrganizationId    string   `json:"organizationId,omitempty"`
}


type PropertyPolicy struct {
	Property *Property `json:"Property"`
	PolicyArray []Policy 
}

type Policies struct{
	PolicyArray []Policy
}

type PolicyRichQuery struct {
	Key string `json:"Key"`
	Record	Policy `json:"Record"`
}
//How to Write Structures
//https://play.golang.org/p/4bpG0FkajS

// ============================================================================================================================
// Main - start the chaincode for Property management
// ============================================================================================================================
func main() {
	err := shim.Start(new(ManageProperties))
	if err != nil {
		fmt.Printf("Error starting Property management chaincode: %s", err)
	}
}

// ============================================================================================================================
// Init - reset all the things
// ============================================================================================================================
func (t *ManageProperties) Init(stub shim.ChaincodeStubInterface) sc.Response {

	fmt.Println("smartProperty Init")
	return shim.Success(nil)

}

// ============================================================================================================================
// Invoke - Our entry Propertyint for Invocations
// ============================================================================================================================
func (t *ManageProperties) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("invoke is running")

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	// function == "init" { //initialize the chaincode state, used as reset
	//return t.Init(stub)
	//else
	if function == "createPolicy" {
		return t.createPolicy(stub, args)
	} else if function == "getPropertybyId" { //Read a Property
		return t.getPropertybyId(stub, args)
	} else if function == "queryProperty" { //Rich Queries
		return t.queryProperty(stub, args)
	} else if function == "getPolicybyId" {
		return t.getPolicybyId(stub, args)
	} else if function == "getHistorybyPropertyId" { //Read complete history of a property
		return t.getHistorybyPropertyId(stub, args)
	}else if function == "getPropertyPolicies" { //Read complete history of a property
		return t.getPropertyPolicies(stub, args)
	}else if function == "queryPoliciesArray" { //Read complete history of a property
		return t.queryPoliciesArray(stub, args)
	}
	

	fmt.Println("invoke did not find func: " + function)
	errMsg := "{ \"message\" : \"Received unknown function invocation\", \"code\" : \"503\"}"
	err := stub.SetEvent("errEvent", []byte(errMsg))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Error("Invalid Smart Contract function name.")
}

// ============================================================================================================================
//  getPropertybyId- get details of a Property from chaincode state
// ============================================================================================================================
func (t *ManageProperties) getPropertybyId(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var jsonResp, errResp string
	fmt.Println("start getPropertybyId")
	fmt.Println("start getPropertybyId", args[0])
	var err error
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 1\" \" as an argument\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}
	_propertyId := args[0]

	jsonResp = ""
	valueAsBytes, err := stub.GetState(_propertyId)
	if err != nil {
		errResp = "{\"Error\":\"Failed to get state for " + _propertyId + "\"}"
		return shim.Error(errResp)
	} else if valueAsBytes == nil {
		fmt.Println(_propertyId + " not found")
		errMsg := "{ \"Property\" : \"" + _propertyId + "\",\"message\" : \"" + "Property Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		jsonResp = string(valueAsBytes[:])
	}
	fmt.Println("jsonResp : " + jsonResp)
	fmt.Println("end getPropertybyId")
	return shim.Success([]byte(jsonResp))
}

// ============================================================================================================================
//  getPolicybyId- get details of all Policy from chaincode state
// ============================================================================================================================
func (t *ManageProperties) getPolicybyId(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var jsonResp, errResp string
	fmt.Println("start getPolicybyId")
	var err error
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 1\" \" as an argument\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}

	_policyId := args[0]

	jsonResp = ""
	valueAsBytes, err := stub.GetState(_policyId)
	if err != nil {
		errResp = "{\"Error\":\"Failed to get state for " + _policyId + "\"}"
		return shim.Error(errResp)
	} else if valueAsBytes == nil {
		fmt.Println(_policyId + " not found")
		errMsg := "{ \"PolicyId\" : \"" + _policyId + "\",\"message\" : \"" + "Policy Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		jsonResp = string(valueAsBytes[:])
	}
	fmt.Println("jsonResp : " + jsonResp)
	fmt.Println("end getPolicybyId")
	return shim.Success([]byte(jsonResp))
}

// ============================================================================================================================
// CreatePolicy - create policy with new or updated policy, store into chaincode state
// ============================================================================================================================
func (t *ManageProperties) createPolicy(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var err error
	var PropertyAsBytes []byte
	var tosend, _propertyId, _policyId string
	var propertyIdArray, policyIdArray []string

	fmt.Println("start createPolicy--------------")
	res := Property{}
	_tempProperty := IntermediateProperty{}

	for _, argsValue := range args {

		err = json.Unmarshal([]byte(argsValue), &_tempProperty)
		if err != nil {
			fmt.Println("Unmarshal args error: ", err)
			return shim.Error("Error unmarshaling arguments: ")
		}
		_propertyId = _tempProperty.PropertyId
		fmt.Println("_propertyId: " + _propertyId)
		PropertyAsBytes, err = stub.GetState(_propertyId)
		if err != nil {
			return shim.Error("Failed to get _propertyId " + _propertyId)
		} else if PropertyAsBytes == nil {
			res.AddressLine1 = _tempProperty.AddressLine1
			res.AddressLine2 = _tempProperty.AddressLine2
			res.City = _tempProperty.City
			res.County = _tempProperty.County
			res.OrganizationName = _tempProperty.OrganizationName
			res.PropertyId = _tempProperty.PropertyId
			res.RelatedPropertyId = _tempProperty.PropertyId
			res.State = _tempProperty.State
			res.TaxID = _tempProperty.TaxID
			res.Zip = _tempProperty.Zip
			res.OrganizationId = _tempProperty.OrganizationId
			res.ObjectType = "Property"
			order, _ := json.Marshal(res)
			fmt.Println("order: " + string(order))
			err = stub.PutState(_propertyId, order) //store Property with PropertyId as key
			if err != nil {
				return shim.Error(err.Error())
			}
			propertyIdArray = append(propertyIdArray, _propertyId)
		}
		for _, policyValue := range _tempProperty.Policies {
			_policyId = policyValue.PolicyId
			_, err = stub.GetState(_policyId)

			if err != nil {
				return shim.Error("Failed to get _policyId " + _policyId)
			}
			newPolicy := Policy{}
			newPolicy.EffectiveDate = policyValue.EffectiveDate
			newPolicy.EffectiveDateText = policyValue.EffectiveDateText
			newPolicy.Exceptions = policyValue.Exceptions
			newPolicy.InsuredNames = policyValue.InsuredNames
			newPolicy.IssueDate = policyValue.IssueDate
			newPolicy.IssuingCompanyName = policyValue.IssuingCompanyName
			newPolicy.LegalDescription = policyValue.LegalDescription
			newPolicy.LiabilityAmount = policyValue.LiabilityAmount
			newPolicy.ObjectType = "Policy"
			newPolicy.OrganizationName = policyValue.OrganizationName
			newPolicy.PropertyId = policyValue.PropertyId
			newPolicy.PolicyId = policyValue.PolicyId
			newPolicy.PolicyNumber = policyValue.PolicyNumber
			newPolicy.TitleVesting = policyValue.TitleVesting
			newPolicy.OrganizationId = policyValue.OrganizationId
			newPolicy.Estate = policyValue.Estate
			newPolicy.Subordinate = policyValue.Subordinate
			newPolicy.ClaimFlag = policyValue.ClaimFlag
			newPolicy.InsuredMortgage = policyValue.InsuredMortgage
			newPolicy.Endorsements = policyValue.Endorsements
			newPolicy.PolicyType = policyValue.PolicyType
			newPolicy.PolicyKingdom = policyValue.PolicyKingdom
			newPolicy.PolicyName = policyValue.PolicyName
			newPolicy.PolicyImage = policyValue.PolicyImage
			policyJSON, _ := json.Marshal(newPolicy)
			fmt.Println("policyJSON: " + string(policyJSON))
			err = stub.PutState(newPolicy.PolicyId, policyJSON) //store Policy with PolicyId as key
			if err != nil {
				return shim.Error(err.Error())
			}
			policyIdArray = append(policyIdArray, newPolicy.PolicyId)
		}
	}
	tosend = "{  \"message\" : \"Polices stored successfully on legder. Please find below details for more information.\",\"Policies successfully stored on ledger with policyIds \" : \"" + strings.Join(policyIdArray, ",") + "\", \"Properties successfully created/updated with propertyIds\" : \"" + strings.Join(propertyIdArray, ",") + "\", \"code\" : \"200\"}"
	fmt.Println("event message: " + tosend)
	err = stub.SetEvent("evtsender", []byte(tosend))
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("end createPolicy")

	return shim.Success(nil)
}



// get property and its all policies 
func (t *ManageProperties) getPropertyPolicies(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	var propertyJsonResp, errResp string
	Property := &Property{}
	var m = &PolicyRichQuery{}
	var ms = []*PolicyRichQuery{}
	fmt.Println("start getPropertyPolicies method")
	var err error
	if len(args) != 1 {
		errMsg := "{ \"message\" : \"Incorrect number of arguments. Expecting 1\" \" as an argument\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	}

	_propertyId := args[0]
	propertyValueAsBytes, err := stub.GetState(_propertyId)
	if err != nil {
		errResp = "{\"Error\":\"Failed to get state for " + _propertyId + "\"}"
		return shim.Error(errResp)
	} else if propertyValueAsBytes == nil {
		fmt.Println(_propertyId + " not found")
		errMsg := "{ \"_propertyId\" : \"" + _propertyId + "\",\"message\" : \"" + "Property Not Found.\", \"code\" : \"503\"}"
		err = stub.SetEvent("errEvent", []byte(errMsg))
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		propertyJsonResp = string(propertyValueAsBytes[:])
	}
	policyQueryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Policy\",\"propertyId\":\"%s\"}}", _propertyId)
	policyQueryResults, err := getQueryResultForQueryString(stub, policyQueryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	fmt.Println("policyQueryResults : " + string(policyQueryResults))
	fmt.Println("propertyJsonResp : " + propertyJsonResp)
	propertyPolicy := &PropertyPolicy{}
	// converting property result to struct
	errProperty := json.Unmarshal([]byte(propertyJsonResp), Property)
	if errProperty != nil {
		return shim.Error(errProperty.Error())
	}
	// converting policies rich query result to struct
	errPolicy := json.Unmarshal([]byte(string(policyQueryResults)), &m)
  	if errPolicy != nil {
    	json.Unmarshal([]byte(string(policyQueryResults)), &ms)
	}
	for _, m := range ms {
		s3, _ := json.Marshal(m.Record)
		fmt.Println(string(s3))
		propertyPolicy.PolicyArray = append(propertyPolicy.PolicyArray, m.Record)
	}
	propertyPolicy.Property = Property
	endResult, _ := json.Marshal(propertyPolicy)
	//endResult = string(endResult)
	fmt.Println("end getPropertyPolicies")
	return shim.Success([]byte(endResult))
}

// =================================================================================================================================
// Get History for a particular property - get history of records for a particular property stored on ledger over the period of time
// =================================================================================================================================
func (t *ManageProperties) getHistorybyPropertyId(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	_propertyId := args[0]

	fmt.Printf("- start getHistorybyPropertyId: %s\n", _propertyId)

	resultsIterator, err := stub.GetHistoryForKey(_propertyId)
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
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
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

	fmt.Printf("- getHistorybyPropertyId returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// =====================================================
// Get Property records filtered by CreatingOrganization
// =====================================================
func (t *ManageProperties) queryPropertyByOwner(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//   0
	// "bob"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	owner := strings.TrimSpace(args[0])

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"Property\",\"creatingOrganization\":\"%s\"}}", owner)

	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// =================================================================
// Get Property records filtered by Selector query sent as arguments
// =================================================================
func (t *ManageProperties) queryPoliciesArray(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	var m = &PolicyRichQuery{}
	var ms = []*PolicyRichQuery{}
	policies := &Policies{}
	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]


	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	errPolicy := json.Unmarshal([]byte(string(queryResults)), &m)
  	if errPolicy != nil {
    	json.Unmarshal([]byte(string(queryResults)), &ms)
	}
	for _, m := range ms {
		s3, _ := json.Marshal(m.Record)
		fmt.Println(string(s3))
		policies.PolicyArray = append(policies.PolicyArray, m.Record)
	}
	endResult, _ := json.Marshal(policies)
	//endResult = string(endResult)
	fmt.Println("end getPropertyPolicies")
	return shim.Success([]byte(endResult))
}


func (t *ManageProperties) queryProperty(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	//   0
	// "queryString"
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]


	queryResults, err := getQueryResultForQueryString(stub, queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// func (t *ManageProperties) queryPropertyWithPagination(stub shim.ChaincodeStubInterface, args []string) sc.Response {

// 	//   0
// 	// "queryString"
// 	if len(args) < 3 {
// 		return shim.Error("Incorrect number of arguments. Expecting 3")
// 	}

// 	queryString := args[0]

// 	pageSize, err := strconv.ParseInt(args[1],10,32)
// 	if err != nil {
// 			return shim.Error(err.Error())
// 	}
// 	bookmark := args[2]

// 	queryResults, err := getQueryResultForQueryStringWithPagination(stub, queryString, int32(pageSize), bookmark)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}
// 	return shim.Success(queryResults)
// }



// func getQueryResultForQueryStringWithPagination(stub shim.ChaincodeStubInterface, queryString string, pageSize int32, bookmark string) ([]byte, error) {

// 	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

// 	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryString, pageSize, bookmark)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resultsIterator.Close()

// 	// buffer is a JSON array containing QueryRecords
// 	buffer, err := constructQueryResponseFromIterator(resultsIterator)
// 	if err != nil {
// 		return nil, err
// 	}

// 	bufferWithPaginationInfo := addPaginationMetadataToQueryResults(buffer, responseMetadata)

// 	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", bufferWithPaginationInfo.String())

// 	return buffer.Bytes(), nil
// }

// func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
// 	// buffer is a JSON array containing QueryResults
// 	var buffer bytes.Buffer
// 	buffer.WriteString("[")

// 	bArrayMemberAlreadyWritten := false
// 	for resultsIterator.HasNext() {
// 		queryResponse, err := resultsIterator.Next()
// 		if err != nil {
// 			return nil, err
// 		}
// 		// Add a comma before array members, suppress it for the first array member
// 		if bArrayMemberAlreadyWritten == true {
// 			buffer.WriteString(",")
// 		}
// 		buffer.WriteString("{\"Key\":")
// 		buffer.WriteString("\"")
// 		buffer.WriteString(queryResponse.Key)
// 		buffer.WriteString("\"")

// 		buffer.WriteString(", \"Record\":")
// 		// Record is a JSON object, so we write as-is
// 		buffer.WriteString(string(queryResponse.Value))
// 		buffer.WriteString("}")
// 		bArrayMemberAlreadyWritten = true
// 	}
// 	buffer.WriteString("]")

// 	return &buffer, nil
// }

// // ===========================================================================================
// // addPaginationMetadataToQueryResults adds QueryResponseMetadata, which contains pagination
// // info, to the constructed query results
// // ===========================================================================================
// func addPaginationMetadataToQueryResults(buffer *bytes.Buffer, responseMetadata *sc.QueryResponseMetadata) *bytes.Buffer {

// 	buffer.WriteString("[{\"ResponseMetadata\":{\"RecordsCount\":")
// 	buffer.WriteString("\"")
// 	buffer.WriteString(fmt.Sprintf("%v", responseMetadata.FetchedRecordsCount))
// 	buffer.WriteString("\"")
// 	buffer.WriteString(", \"Bookmark\":")
// 	buffer.WriteString("\"")
// 	buffer.WriteString(responseMetadata.Bookmark)
// 	buffer.WriteString("\"}}]")

// 	return buffer
// }


// =========================================================================================
// Helper function for queryProperty functionality.
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
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

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}
