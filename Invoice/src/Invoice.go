package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strconv"

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
	In_MaterialNumber string `json:"in_MaterialNumber"`
	In_PurchaseOrderNumber string `json:"in_PurchaseOrderNumber"`
	InvoiceAmount string `json:"invoiceAmount"`
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

/*
 * Function to create invoice and store invoice amount onto blockchain for a specific purchase order and material number.
 */
func (cc *Invoice) createInvoice(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	// check total parameters
	if len(args) != 4 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if Invoice already exists
	if validateValue, validateErr := stub.GetState("IN-"+args[0]+"-"+args[1]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Invoice already exists")
	}

	// create json for invoice object
	info := &Invoice{
		In_MaterialNumber: args[0],
		In_PurchaseOrderNumber: args[1],
		InvoiceAmount: args[2],
	}

	// convert to byte
	jsonInvoiceInfo, _ := json.Marshal(info)

	// write Invoice and details to BC
	if err := stub.PutState("IN-"+args[0]+"-"+args[1], jsonInvoiceInfo); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Invoice Created Successsfully!", nil)

}

/**
 * Function to get Invoice amount by purchase order id and material number
 */
func (cc *Invoice) getInvoiceAmountById(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 1st - purchase order #,  2nd - MaterialNumber, 3rd - expected date, 4th - actual date
	
	// check total parameters
	if len(args) != 4 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// query for bc to fetch all the blocks where purchase order number = arsg[0] and material number  = args[1]
	queryString := fmt.Sprintf("{\"selector\":{\"in_PurchaseOrderNumber\":\""+args[0]+"\",\"in_MaterialNumber\":\""+args[1]+"\"}}")
	
	// invoke query
	resultsIterator, err := stub.GetQueryResult(queryString)

	// if error, return error as response
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	// if no error, close the iterator and return
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	
	// iterate over all the blocks of data obtained as a result from the query
	for resultsIterator.HasNext() {

		// fetch idividual block
		queryResponse, err1 := resultsIterator.Next()

		// if error fetching individual block, return error in response
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		
		// define a var for Invoice json structure
		var invoiceData Invoice

		// convert bytes to json
		json.Unmarshal(queryResponse.Value,&invoiceData)

		// create invoice object
		buffer = createInvoiceObject(invoiceData,args[2],args[3],buffer)

		// because only latest block for purchase order and material number needs to be sent as the response
		break;
	}
	
	// return bytes with success status
	return Success(200, "OK", buffer.Bytes())
}

/**
 * Function to create invoice object which will be sent as the response in form of bytes
 */
func createInvoiceObject(invoiceData Invoice,expectedDate string,actualDate string,buffer bytes.Buffer) (x bytes.Buffer) {
	// store invoice amount in buffer
	buffer.WriteString("{\"invoiceAmount\":")
	buffer.WriteString("\"")
	buffer.WriteString(invoiceData.InvoiceAmount)
	buffer.WriteString("\"")

	// time format
	timeFormat := "01/02/2006"

	// get current date - it should come from request parameter in actual prod version, because different nodes could be 
	//geographically located in different region and this transaction could be rejected by validating nodes
	currDate := time.Now()
	
	// convert curr date into string
	currDateInStr := x.Format(timeFormat)

	// variables to keep track of status, state and penalty amount
	status := "", state := "", delayPenalty := ""

	// case 1: when expected date is not empty but actual date is empty
	if expectedDate != "" && actualDate == "" {
		// get difference between current date and expected date
		diff = := currDateInStr.sub(expectedDate)

		// get diff in terms of days
		dayDiff := (diff.Hours())/24

		// case 1.1: when day diff is great then 0, then material delivery is delayed
		if dayDiff > float64(0) {
			// diff between curr date and expected is > 0 and actual date is not present, delayed + diff between currDate and ExpectedDate
			x := fmt.Sprintf("%.0f",dayDiff)
			status = "Delayed+"x
			state = "Error"
			delayPenalty = "-"
		} else {	// case 1.2: if expected date is greater then current date, it is assumed material will be delivered on time.
			// on time
			status = "On-Time"
			state = "Success"
			delayPenalty = "0.00"
		}

	} else if expectedDate != "" && actualDate != "" { // case 2: when expected and actual date are present
		// parsing expected date
		ed, _ := time.Parse(timeFormat,expectedRawMaterialInfo.ExpectedDate)

		// parsing actual date
		ad, _ := time.Parse(timeFormat,actualDate)
		
		// calculate diff between expected and actual date
		duration := ad.Sub(ed)

		// diff in days
		days := (duration.Hours())/24

		// case 2.1: when diff in days is less then or equal to 0, it is considered that material is delivered to manufacturer with 0 penalty
		if days <= float64(0) {
			// delivered
			status = "Delivered"
			state = "None"
			delayPenalty = "0.00"
		} else { // case 2.2: when diff in days is greater than zero, it is considered that material is delivered but with diff in days delay and penalty will be incurred
			y := fmt.Sprintf("%.0f",days)
			status = "Delivered+"+y
			state = "Error"

			// current penalty percentage
			penaltyPercentage := 0
					
			// assign penalty pecentage based on diff in # of days
			if 	days > float64(0) && days <= float64(2) {
				penaltyPercentage = 5
			}else if days > float64(2) && days <= float64(7) {
				penaltyPercentage = 10
			}else if days > float64(7) {
				penaltyPercentage = 20
			}
			
			// calculate penalty amount
			InvoiceAmountFloat, _ := strconv.ParseFloat((invoice.InvoiceAmount),32)
			InvoicePenalty := (InvoiceAmountFloat*float64(penaltyPercentage)) / 100
			invoicePenaltyStr := fmt.Sprintf("%.2f", InvoicePenalty );

			delayPenalty = invoicePenaltyStr
		}

	} 

	// store status in buffer
	buffer.WriteString(",\"status\":")
	buffer.WriteString("\"")
	buffer.WriteString(status)
	buffer.WriteString("\"")

	// store state in buffer
	buffer.WriteString(",\"state\":")
	buffer.WriteString("\"")
	buffer.WriteString(state)
	buffer.WriteString("\"")

	// store delayPenalty in buffer
	buffer.WriteString(",\"delayPenalty\":")
	buffer.WriteString("\"")
	buffer.WriteString(delayPenalty)
	buffer.WriteString("\"}")

	x = buffer
	return
}