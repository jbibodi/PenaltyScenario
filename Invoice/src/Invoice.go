package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

func Success(rc int32, message string, payload []byte) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: message,
		Payload: payload,
	}
}

func Error(rc int32, message string) peer.Response {
	return peer.Response{
		Status:  rc,
		Message: message,
	}
}

func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}


type Invoice struct {
	InvoiceNumber string `json:"invoiceNumber"`	
	RawMaterialNumber string `json:"rawMaterialNumber"`
	DemandNumber string `json:"demandNumber"`
	InvoiceAmount string `json:"invoiceAmount"`
	InvoiceInfo string `json:"invoiceInfo"`
}

func main() {
	if err := shim.Start(new(Invoice)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
}

// Init is called during Instantiate transaction.
func (cc *Invoice) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return Success(http.StatusOK, "OK", nil)
}

// Invoke is called to update or query the blockchain
func (cc *Invoice) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()
	// Route call to the correct function
	switch function {
		case "createInvoice":
			return cc.createInvoice(stub, args)	
		case "getInvoiceAmountById":
			return cc.getInvoiceAmountById(stub, args)
		default:
			return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'createInvoice|getInvoiceAmountById'!")
	}
}

func (cc *Invoice) createInvoice(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 4 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if Invoice already exists
	if validateValue, validateErr := stub.GetState("IN-"+args[0]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Invoice already exists")
	}

	info := &Invoice{
		InvoiceNumber: args[0],
		RawMaterialNumber: args[1],
		DemandNumber: args[2],
		InvoiceAmount: args[3],
		InvoiceInfo: "true",
	}

	// convert to byte
	jsonInvoiceInfo, _ := json.Marshal(info)

	// write Invoice and details to BC
	if err := stub.PutState("IN-"+args[0], jsonInvoiceInfo); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Invoice Created Successsfully!", nil)

}

func (cc *Invoice) getInvoiceAmountById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 1st - demand # and 2nd - rawMaterialNumber
	
	if len(args) != 2 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	queryString := fmt.Sprintf("{\"selector\":{\"invoiceInfo\":\"true\",\"demandNumber\":\""+args[0]+"\",\"rawMaterialNumber\":\""+args[1]+"\"}}")
	
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	
	for resultsIterator.HasNext() {
		queryResponse, err1 := resultsIterator.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		
		var invoiceData Invoice
		json.Unmarshal(queryResponse.Value,&invoiceData)
		
		buffer.WriteString("{\"invoiceAmount\":")
		buffer.WriteString("\"")
		buffer.WriteString(invoiceData.InvoiceAmount)
		buffer.WriteString("\"}")
		
		break;
	}
	
	return Success(200, "OK", buffer.Bytes())
}
