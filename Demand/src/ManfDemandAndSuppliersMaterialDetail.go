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

type PurchaseOrder struct {
	PurchaseOrderNumber string `json:"purchaseOrderNumber"`
	SupplierCode string `json:"supplierCode"`
	SupplierLocation string `json:"supplierLocation"`
	IsPurchaseOrderObject bool `json:"isPurchaseOrderObject"`
}

type ExpectedMaterialInformation struct {
	MaterialNumber string `json:"materialNumber"`
	Ex_PurchaseOrderNumber string `json:"ex_PurchaseOrderNumber"`
	ExpectedDate string `json:"expectedDate"`
	IsExpectedMaterialInfoObject bool `json:"isExpectedMaterialInfoObject"`
}

type ActualMaterialInformation struct {
	MaterialNumber string `json:"materialNumber"`
	Ac_PurchaseOrderNumber string `json:"ac_PurchaseOrderNumber"`
	DelayReason string `json:"delayReason"`
	ActualDate string `json:"actualDate"`
}
	
type Invoice struct {
	InvoiceAmount float64 `json:"invoiceAmount"`
	DelayPenalty float64 `json:"delayPenalty"`
}

type TrackOrder struct {
	TrackingId string `json:"trackingId"`
	TrackMaterialNumber string `json:"trackMaterialNumber"`
	TrackPurchaseOrderNumber string `json:"trackPurchaseOrderNumber"`
	SupplierFacilityName string `json:"supplierFacilityName"`
	TrackStatus string `json:"trackStatus"`
	Timestamp string `json:"timestamp"`
	TrackOrderReason string `json:"trackOrderReason"`
	TrackOrderState string `json:"trackOrderState"`
}

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

func main() {
	if err := shim.Start(new(PurchaseOrder)); err != nil {
		fmt.Printf("Main: Error starting chaincode: %s", err)
	}
}

func (cc *PurchaseOrder) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return Success(http.StatusOK, "OK", nil)
}

func (cc *PurchaseOrder) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()
	
	// Route call to the correct function
	switch function {
		case "createPurchaseOrder":
			return cc.createPurchaseOrder(stub, args)
		case "createExpectedMaterialInformation":
			return cc.createExpectedMaterialInformation(stub,args)
		case "createActualMaterialInformation":
			return cc.createActualMaterialInformation(stub,args)
		case "createMaterialTracking":
			return cc.createMaterialTracking(stub,args)
		case "getAllPurchaseOrder":
			return cc.getAllPurchaseOrder(stub,args)
		case "getAllMaterialInformation":
			return cc.getAllMaterialInformation(stub,args)
		default:
			return Error(http.StatusNotImplemented, "Invalid method! Valid methods are 'createPurchaseOrder|createExpectedMaterialInformation|createActualMaterialInformation|getAllPurchaseOrder|getAllMaterialInformation|createMaterialTracking'!")
	}
}

func (cc *PurchaseOrder) createPurchaseOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if purchase order already exists
	if validateValue, validateErr := stub.GetState(args[0]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Purchase order already exists")
	}
	
	purchaseOrderObject := &PurchaseOrder{
		PurchaseOrderNumber: args[0],
		SupplierCode: args[1],
		SupplierLocation: args[2],
		isPurchaseOrderObject: true,
	}	

	// convert to byte
	purchaseOrderObjectInBytes, _ := json.Marshal(purchaseOrderObject)

	// write PurchaseOrder and details to BC
	if err := stub.PutState(args[0], purchaseOrderObjectInBytes); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Purchase Order Created Successsfully!", nil)
}

func (cc *PurchaseOrder) createExpectedMaterialInformation(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 3 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if material expected date info already exists
	if validateValue, validateErr := stub.GetState("Ex-"+args[0]+"-"+args[1]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Expected Date for purchase order "+args[1]+" and material number "+args[0]+" already exists!")
	}
	
	expectedMaterialInformationObject := &ExpectedMaterialInformation{
		MaterialNumber: args[0],
		Ex_PurchaseOrderNumber: args[1],
		ExpectedDate: args[2],
		IsExpectedMaterialInfoObject: true,
	}	

	// convert to byte
	expectedMaterialInformationObjectInBytes, _ := json.Marshal(expectedMaterialInformationObject)

	// write material info to BC
	if err := stub.PutState("Ex-"+args[0]+"-"+args[1], expectedMaterialInformationObjectInBytes); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Material's expected delivery date information created successsfully!", nil)
}

func (cc *PurchaseOrder) createActualMaterialInformation(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 4 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if raw material expected date info already exists
	if validateValue, validateErr := stub.GetState("Ac-"+args[0]+"-"+args[1]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Actual Date for purchase order "+args[1]+" and material number "+args[0]+" already exists!")
	}
	
	actualMaterialInformationObject := &ActualMaterialInformation{
		MaterialNumber: args[0],
		Ac_PurchaseOrderNumber: args[1],
		ActualDate: args[2],
		DelayReason: args[3],
	}	

	// convert to byte
	actualMaterialInformationObjectInBytes, _ := json.Marshal(actualMaterialInformationObject)

	// write raw material info to BC
	if err := stub.PutState("Ac-"+args[0]+"-"+args[1], actualMaterialInformationObjectInBytes); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Material's actual delivery date information created successsfully!", nil)
}

func (cc *PurchaseOrder) createMaterialTracking(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	
	if len(args) != 8 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	
	// Check if raw material expected date info already exists
	if validateValue, validateErr := stub.GetState("TR-"+args[0]); validateErr != nil || validateValue != nil {
		return Error(http.StatusConflict, "Tracking Number already exists")
	}

	trackOrderObject := &TrackOrder{
		TrackingId: args[0],
		TrackMaterialNumber: args[1],
		TrackPurchaseOrderNumber: args[2],
		SupplierFacilityName: args[3],
		Timestamp: args[4],
		TrackStatus: args[5],
		TrackOrderState: args[6],
		TrackOrderReason: args[7],
	}
	
	// convert to byte
	trackOrderObjectInBytes, _ := json.Marshal(trackOrderObject)

	// write raw material info to BC
	if err := stub.PutState("TR-"+args[0], trackOrderObjectInBytes); err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}

	return Success(http.StatusCreated,"Tracking Information Created Successsfully!", nil)
}


func (cc *PurchaseOrder) getAllPurchaseOrder(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	queryStringToGetAllPurchaseOrder := fmt.Sprintf("{\"selector\":{\"isPurchaseOrderObject\":true}}")
	allPurchaseOrderResults, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer allPurchaseOrderResults.Close()
	
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")
	bArrayMemberAlreadyWritten := false
	
	for allPurchaseOrderResults.HasNext() {
		queryResponse, err1 := allPurchaseOrderResults.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		
		var purchaseOrderObject PurchaseOrder
		json.Unmarshal(queryResponse.Value,&purchaseOrderObject)
		buffer.WriteString("{")
		
		buffer = generatePurchaseOrderObject(purchaseOrderObject,buffer)
		buffer.WriteString(",")
		
		buffer = getMaterialInformation(purchaseOrderObject.PurchaseOrderNumber,buffer)
		
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	
	buffer.WriteString("]}")
	return Success(200, "OK", buffer.Bytes())
}

func generatePurchaseOrderObject(purchaseOrderObject PurchaseOrder,buffer bytes.Buffer) (x bytes.Buffer) {
	buffer.WriteString("\"purchaseOrderNumber\":")
	buffer.WriteString("\"")
	buffer.WriteString(purchaseOrderObject.PurchaseOrderNumber)
	buffer.WriteString("\"")
	
	buffer.WriteString(", \"supplierCode\":")
	buffer.WriteString("\"")
	buffer.WriteString(purchaseOrderObject.SupplierCode)
	buffer.WriteString("\"")
	
	buffer.WriteString(", \"supplierLocation\":")
	buffer.WriteString("\"")
	buffer.WriteString(purchaseOrderObject.SupplierLocation)
	buffer.WriteString("\"")
	
	x = buffer
	return
}

func getMaterialInformation(purchaseOrderNumber string,buffer bytes.Buffer)(x bytes.Buffer) {
	// for material info
	buffer.WriteString("\"expectedRawMaterialInformation\":")
	buffer.WriteString("[")
	
	partInfoAlreadyWritten := false
	queryString = fmt.Sprintf("{\"selector\":{\"isExpectedMaterialInfoObject\":true,\"ex_PurchaseOrderNumber\":\""+purchaseOrderNumber+"\"}}")
	expectedPartResultsIterator, expectedPartErr := stub.GetQueryResult(queryString)
	
	if expectedPartErr != nil {
		return Error(http.StatusInternalServerError, expectedPartErr.Error())
	}
	
	defer expectedPartResultsIterator.Close()
	
	totalDelivered := 0, totalParts := 0, isEmptyActualDate := false, parentExpectedDate := "", parentStatus := "", parentActualDate := "", parentDelayReason := "", currMax := float64(0), parentState := ""
	
	for expectedPartResultsIterator.HasNext() {
		totalParts = totalParts + 1;
		expectedPartResponse, _ := expectedPartResultsIterator.Next()
		
		if partInfoAlreadyWritten == true {
			buffer.WriteString(",")
		}
		
		var expectedMaterialInformation ExpectedMaterialInformation
		json.Unmarshal(expectedPartResponse.Value,&expectedMaterialInformation)
		
		buffer = getMaterialInformation(purchaseOrderNumber,expectedMaterialInformation,buffer)
		
		partInfoAlreadyWritten = true
	}
	
	x = buffer
	return
}

func getMaterialInformation(purchaseOrderNumber string,expectedMaterialInformation ExpectedMaterialInformation,buffer bytes.Buffer) (x bytes.Buffer) {
	buffer.WriteString("{\"rawMaterialNumber\":")
	buffer.WriteString("\"")
	buffer.WriteString(expectedMaterialInformation.MaterialNumber)
	buffer.WriteString("\"")
	buffer.WriteString(",")

	buffer = getTrackingInfo(purchaseOrderNumber,expectedMaterialInformation.MaterialNumber,buffer)
	buffer.WriteString(",")

	actualDate := ""

	buffer,actualDate = getActualDateInformationForMaterial(purchaseOrderNumber,expectedMaterialInformation.MaterialNumber,buffer)
	buffer.WriteString(",")

	buffer = getInvoiceInformation(purchaseOrderNumber,expectedMaterialInformation,actualDate,buffer)
	buffer.WriteString(",")

	x = buffer
	return
}

func getInvoiceInformation(purchaseOrderNumber string,expectedMaterialInformation ExpectedMaterialInformation,actualDate string,buffer bytes.Buffer) (x bytes.Buffer) {
	// Check if invoice exists
	f := "getInvoiceAmountById"
	queryArgs := toChaincodeArgs(f, purchaseOrderNumber,expectedMaterialInformation.MaterialNumber,expectedMaterialInformation.ExpectedDate,actualDate)

	invoiceResponse := stub.InvokeChaincode("a3596e82-9760-494a-bad7-31ffc9530b7e-com-sap-icn-blockchain-invoice-penalty-scenario",queryArgs,"")	

	var invoice Invoice
	json.Unmarshal(invoiceResponse.Payload, &invoice)
	
	buffer.WriteString("\"invoiceAmount\":")
	buffer.WriteString("\"")
	buffer.WriteString(invoice.InvoiceAmount)
	buffer.WriteString("\"")
	
	buffer.WriteString("\"delayPenalty\":")
	buffer.WriteString("\"")
	buffer.WriteString(invoice.DelayPenalty)
	buffer.WriteString("\"")

	timeFormat := "01/02/2006"
	stateOfDelivery := ""
}

func getActualDateInformationForMaterial(purchaseOrderNumber string, materialNumber string,buffer bytes.Buffer) (x bytes.Buffer,actualDate string,) {
	queryString = fmt.Sprintf("{\"selector\":{\"ac_PurchaseOrderNumber\":\""+purchaseOrderNumber+"\",\"materialNumber\":\""+materialNumber+"\"}}")
			
	actualPartResultsIterator, _ := stub.GetQueryResult(queryString)
	defer actualPartResultsIterator.Close()
	
	actualDate := ""
	delayReasonTemp := ""
	
	for actualPartResultsIterator.HasNext() {
		actualPartResponse, _ := actualPartResultsIterator.Next()
		
		var actualMaterialInfo ActualMaterialInformation
		json.Unmarshal(actualPartResponse.Value,&actualMaterialInfo)
		
		buffer.WriteString(", \"delayReason\":")
		buffer.WriteString("\"")
		buffer.WriteString(actualMaterialInfo.DelayReason)
		buffer.WriteString("\"")
		delayReasonTemp = actualMaterialInfo.DelayReason
		
		buffer.WriteString(", \"actualDate\":")
		buffer.WriteString("\"")
		buffer.WriteString(actualMaterialInfo.ActualDate)
		buffer.WriteString("\"")
		
		actualDate = actualMaterialInfo.ActualDate
		
		break;
	}
	x = buffer
	return
}

func getTrackingInfo(purchaseOrderNumber string, materialNumber string,buffer bytes.Buffer) (x bytes.Buffer) {
	queryString = fmt.Sprintf("{\"selector\":{\"trackPurchaseOrderNumber\":\""+purchaseOrderNumber+"\",\"trackMaterialNumber\":\""+materialNumber+"\"}}")
	trackingPartResultsIterator, _ := stub.GetQueryResult(queryString)
	defer trackingPartResultsIterator.Close()
	
	buffer.WriteString("\"trackingInfo\":")
	buffer.WriteString("[")
	
	isTrackingInfoPresent := false
	
	for trackingPartResultsIterator.HasNext() {
		trackingPartResponse, _ := trackingPartResultsIterator.Next()
		
		var trackingInfo TrackOrder
		json.Unmarshal(trackingPartResponse.Value,&trackingInfo)
		
		if isTrackingInfoPresent == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"supplierFacilityName\":")
		buffer.WriteString("\"")
		buffer.WriteString(trackingInfo.SupplierFacilityName)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"trackStatus\":")
		buffer.WriteString("\"")
		buffer.WriteString(trackingInfo.TrackStatus)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"reason\":")
		buffer.WriteString("\"")
		buffer.WriteString(trackingInfo.TrackOrderReason)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"state\":")
		buffer.WriteString("\"")
		buffer.WriteString(trackingInfo.TrackOrderState)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(trackingInfo.Timestamp)
		buffer.WriteString("\"}")
		
		isTrackingInfoPresent = true
	}
	
	buffer.WriteString("]")
	x = buffer
	return
}






func (cc *Demand) getAllDemand(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	queryString := fmt.Sprintf("{\"selector\":{\"demandInfo\":\"true\"}}")
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")

	bArrayMemberAlreadyWritten := false
	
	for resultsIterator.HasNext() {
		queryResponse, err1 := resultsIterator.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		
		var demandInfoObject Demand
		json.Unmarshal(queryResponse.Value,&demandInfoObject)
		
		buffer.WriteString("{\"demandNumber\":")
		buffer.WriteString("\"")
		buffer.WriteString(demandInfoObject.DemandNumber)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"supplierCode\":")
		buffer.WriteString("\"")
		buffer.WriteString(demandInfoObject.SupplierCode)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"supplierLocation\":")
		buffer.WriteString("\"")
		buffer.WriteString(demandInfoObject.SupplierLocation)
		buffer.WriteString("\"")
		
		// for raw material info
		buffer.WriteString(", \"expectedRawMaterialInformation\":")
		buffer.WriteString("[")
		
		partInfoAlreadyWritten := false
		queryString = fmt.Sprintf("{\"selector\":{\"rawMaterialExpectedInfo\":\"true\",\"demandNumber\":\""+(demandInfoObject.DemandNumber)+"\"}}")
		expectedPartResultsIterator, expectedPartErr := stub.GetQueryResult(queryString)
		if expectedPartErr != nil {
			return Error(http.StatusInternalServerError, expectedPartErr.Error())
		}
		defer expectedPartResultsIterator.Close()
		
		totalDelivered := 0 
		totalParts := 0
		isEmptyActualDate := false
		parentExpectedDate := ""	
		
		parentStatus := ""
		parentActualDate := ""
		parentDelayReason := ""
		currMax := float64(0)
		parentState := ""
		
		for expectedPartResultsIterator.HasNext() {
			totalParts = totalParts + 1;
			expectedPartResponse, _ := expectedPartResultsIterator.Next()
			
			if partInfoAlreadyWritten == true {
				buffer.WriteString(",")
			}
			
			var expectedRawMaterialInfo RawMaterialExpectedInformation
			json.Unmarshal(expectedPartResponse.Value,&expectedRawMaterialInfo)
			
			if parentExpectedDate == "" {
				parentExpectedDate = expectedRawMaterialInfo.ExpectedDate
			}
			
			buffer.WriteString("{\"rawMaterialNumber\":")
			buffer.WriteString("\"")
			buffer.WriteString(expectedRawMaterialInfo.RawMaterialNumber)
			buffer.WriteString("\"")
			
			// fetch tracking
			queryString = fmt.Sprintf("{\"selector\":{\"trackOrderNumber\":\""+(demandInfoObject.DemandNumber)+"\",\"trackPartNumber\":\""+(expectedRawMaterialInfo.RawMaterialNumber)+"\"}}")
			trackingPartResultsIterator, _ := stub.GetQueryResult(queryString)
			defer trackingPartResultsIterator.Close()
			
			buffer.WriteString(", \"trackingInfo\":")
			buffer.WriteString("[")
			
			isTrackingInfoPresent := false
			
			for trackingPartResultsIterator.HasNext() {
				trackingPartResponse, _ := trackingPartResultsIterator.Next()
				
				var trackingInfo TrackOrder
				json.Unmarshal(trackingPartResponse.Value,&trackingInfo)
				
				if isTrackingInfoPresent == true {
					buffer.WriteString(",")
				}
				
				buffer.WriteString("{\"supplierFacilityName\":")
				buffer.WriteString("\"")
				buffer.WriteString(trackingInfo.SupplierFacilityName)
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"trackStatus\":")
				buffer.WriteString("\"")
				buffer.WriteString(trackingInfo.TrackStatus)
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"reason\":")
				buffer.WriteString("\"")
				buffer.WriteString(trackingInfo.Reason)
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"state\":")
				buffer.WriteString("\"")
				buffer.WriteString(trackingInfo.State)
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"timestamp\":")
				buffer.WriteString("\"")
				buffer.WriteString(trackingInfo.Timestamp)
				buffer.WriteString("\"}")
				
				isTrackingInfoPresent = true
			}
			
			buffer.WriteString("]")
			
			queryString = fmt.Sprintf("{\"selector\":{\"rawMaterialActualInfo\":\"true\",\"demandNumber\":\""+(demandInfoObject.DemandNumber)+"\",\"rawMaterialNumber\":\""+(expectedRawMaterialInfo.RawMaterialNumber)+"\"}}")
			
			actualPartResultsIterator, _ := stub.GetQueryResult(queryString)
			defer actualPartResultsIterator.Close()
			
			actualDate := ""
			delayReasonTemp := ""
			
			for actualPartResultsIterator.HasNext() {
				actualPartResponse, _ := actualPartResultsIterator.Next()
				
				var actualRawMaterialInfo RawMaterialActualInformation
				json.Unmarshal(actualPartResponse.Value,&actualRawMaterialInfo)
				
				buffer.WriteString(", \"delayReason\":")
				buffer.WriteString("\"")
				buffer.WriteString(actualRawMaterialInfo.DelayReason)
				buffer.WriteString("\"")
				delayReasonTemp = actualRawMaterialInfo.DelayReason
				
				buffer.WriteString(", \"actualDate\":")
				buffer.WriteString("\"")
				buffer.WriteString(actualRawMaterialInfo.ActualDate)
				buffer.WriteString("\"")
				
				actualDate = actualRawMaterialInfo.ActualDate
				
				break;
			}
			
			// Check if invoice exists
			f := "getInvoiceAmountById"
			queryArgs := toChaincodeArgs(f, demandInfoObject.DemandNumber,expectedRawMaterialInfo.RawMaterialNumber)

			invoiceResponse := stub.InvokeChaincode("a3596e82-9760-494a-bad7-31ffc9530b7e-com-sap-icn-blockchain-invoice-penalty-scenario",queryArgs,"")	

			var invoice Invoice
			json.Unmarshal(invoiceResponse.Payload, &invoice)
			
			buffer.WriteString(", \"invoiceAmount\":")
			buffer.WriteString("\"")
			buffer.WriteString(invoice.InvoiceAmount)
			buffer.WriteString("\"")
			
			timeFormat := "01/02/2006"
			stateOfDelivery := ""
			
			if parentActualDate == "" {
				parentActualDate = actualDate
			}
			
			if expectedRawMaterialInfo.ExpectedDate != "" && actualDate != "" {
				totalDelivered = totalDelivered + 1;
				ed, _ := time.Parse(timeFormat,expectedRawMaterialInfo.ExpectedDate)
				ad, _ := time.Parse(timeFormat,actualDate)
				
				duration := ad.Sub(ed)
				days := (duration.Hours())/24
				xy := ""
				
				if days == float64(0) {
					buffer.WriteString(", \"status\":")
					buffer.WriteString("\"")
					buffer.WriteString("Delivered")
					buffer.WriteString("\"")
					xy = "Delivered"
					stateOfDelivery = "None"
					
					buffer.WriteString(", \"state\":")
					buffer.WriteString("\"")
					buffer.WriteString("None")
					buffer.WriteString("\"")
					
					buffer.WriteString(", \"delayPenalty\":")
					buffer.WriteString("\"")
					buffer.WriteString("0.00")
					buffer.WriteString("\"")
				} else {
					buffer.WriteString(", \"status\":")
					buffer.WriteString("\"")
					x := fmt.Sprintf("%.0f",days)
					strForStatus := "Delivered+"+x
					buffer.WriteString(strForStatus)
					buffer.WriteString("\"")
					xy = "Delivered+"+x
					stateOfDelivery = "Error"
					
					buffer.WriteString(", \"state\":")
					buffer.WriteString("\"")
					buffer.WriteString("Error")
					buffer.WriteString("\"")
					
					penaltyPercentage := 0
					
					if 	days > float64(0) && days <= float64(2) {
						penaltyPercentage = 5
					}else if days > float64(2) && days <= float64(7) {
						penaltyPercentage = 10
					}else if days > float64(7) {
						penaltyPercentage = 20
					}
					
					InvoiceAmountFloat, _ := strconv.ParseFloat((invoice.InvoiceAmount),32)
					InvoicePenalty := (InvoiceAmountFloat*float64(penaltyPercentage)) / 100
					invoicePenaltyStr := fmt.Sprintf("%.2f", InvoicePenalty );
					
					buffer.WriteString(", \"delayPenalty\":")
					buffer.WriteString("\"")
					buffer.WriteString(invoicePenaltyStr)
					buffer.WriteString("\"")
					
				}
				
				if parentActualDate != "" {
					edLatest, _ := time.Parse(timeFormat,parentActualDate)
					adLatest, _ := time.Parse(timeFormat,actualDate)
					
					durationLatest := adLatest.Sub(edLatest)
					latestDays := (durationLatest.Hours())/24
					
					if currMax <= latestDays {
						currMax = latestDays
						parentActualDate = actualDate
						parentStatus = xy
						parentDelayReason = delayReasonTemp
						parentState = stateOfDelivery
					}
					
				} else {
					parentActualDate = actualDate
					parentStatus = xy
					parentDelayReason = delayReasonTemp
					parentState = "None"
				}
			} else {
				isEmptyActualDate = true
				buffer.WriteString(", \"delayReason\":")
				buffer.WriteString("\"")
				buffer.WriteString("")
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"actualDate\":")
				buffer.WriteString("\"")
				buffer.WriteString("")
				buffer.WriteString("\"")
			
				buffer.WriteString(", \"status\":")
				buffer.WriteString("\"")
				buffer.WriteString("On-Time")
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"state\":")
				buffer.WriteString("\"")
				buffer.WriteString("Success")
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"delayPenalty\":")
				buffer.WriteString("\"")
				buffer.WriteString("0.00")
				buffer.WriteString("\"")
			}
			
			buffer.WriteString(", \"expectedDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(expectedRawMaterialInfo.ExpectedDate)
			buffer.WriteString("\"}")
			
			partInfoAlreadyWritten = true
			
		}
		
		buffer.WriteString("]")
		
		if isEmptyActualDate == true {		// when actual date is not present or is empty, set status, actual date and delay reason at order level
			parentStatus = "On-Time"
			buffer.WriteString(", \"parentStatus\":")
			buffer.WriteString("\"")
			buffer.WriteString(parentStatus)
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"state\":")
			buffer.WriteString("\"")
			buffer.WriteString("Success")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"actualDate\":")
			buffer.WriteString("\"")
			buffer.WriteString("")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"delayReason\":")
			buffer.WriteString("\"")
			buffer.WriteString("No delay")
			buffer.WriteString("\"")
		} else {							// when actual date is present, set status, actual date and delay reason at order level
			buffer.WriteString(", \"parentStatus\":")
			buffer.WriteString("\"")
			buffer.WriteString(parentStatus)
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"state\":")
			buffer.WriteString("\"")
			buffer.WriteString(parentState)
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"parentActualDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(parentActualDate)
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"delayReason\":")
			buffer.WriteString("\"")
			buffer.WriteString(parentDelayReason)
			buffer.WriteString("\"")
		}
		
		buffer.WriteString(", \"parentExpectedDate\":")
		buffer.WriteString("\"")
		buffer.WriteString(parentExpectedDate)
		buffer.WriteString("\"")
		
		overAllShipmentPercent := (float64(totalDelivered)/float64(totalParts) * 100)
		buffer.WriteString(", \"overAllShipmentStatus\":")
		buffer.WriteString("\"")
		shipmentPercent := fmt.Sprintf("%.2f",overAllShipmentPercent)
		buffer.WriteString(shipmentPercent)
		buffer.WriteString("\"}")
		
		bArrayMemberAlreadyWritten = true
	}
	
	buffer.WriteString("]}")
	return Success(200, "OK", buffer.Bytes())
}

/*func (cc *Demand) getPartByDemandNumber(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return Error(http.StatusNotAcceptable, "Invalid parameters!")
	}
	return shim.Success(nil)
}*/



func (cc *Demand) getAllPartNumber(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	queryString := fmt.Sprintf("{\"selector\":{\"rawMaterialExpectedInfo\":\"true\"}}")
	
	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return Error(http.StatusInternalServerError, err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("{ \"values\": [")

	bArrayMemberAlreadyWritten := false
	
	for resultsIterator.HasNext() {
		queryResponse, err1 := resultsIterator.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}
		
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		
		var expectedRawMaterialInfo RawMaterialExpectedInformation
		json.Unmarshal(queryResponse.Value,&expectedRawMaterialInfo)
		
		buffer.WriteString("{\"rawMaterialNumber\":")
		buffer.WriteString("\"")
		buffer.WriteString(expectedRawMaterialInfo.RawMaterialNumber)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"expectedDate\":")
		buffer.WriteString("\"")
		buffer.WriteString(expectedRawMaterialInfo.ExpectedDate)
		buffer.WriteString("\"")
		
		buffer.WriteString(", \"demandNumber\":")
		buffer.WriteString("\"")
		buffer.WriteString(expectedRawMaterialInfo.DemandNumber)
		buffer.WriteString("\"")
		
		queryString = fmt.Sprintf("{\"selector\":{\"rawMaterialActualInfo\":\"true\",\"demandNumber\":\""+(expectedRawMaterialInfo.DemandNumber)+"\",\"rawMaterialNumber\":\""+(expectedRawMaterialInfo.RawMaterialNumber)+"\"}}")
			
		actualPartResultsIterator, _ := stub.GetQueryResult(queryString)
		defer actualPartResultsIterator.Close()
		
		actualDate := ""
		
		for actualPartResultsIterator.HasNext() {
			actualPartResponse, _ := actualPartResultsIterator.Next()
			
			var actualRawMaterialInfo RawMaterialActualInformation
			json.Unmarshal(actualPartResponse.Value,&actualRawMaterialInfo)
			
			buffer.WriteString(", \"delayReason\":")
			buffer.WriteString("\"")
			buffer.WriteString(actualRawMaterialInfo.DelayReason)
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"actualDate\":")
			buffer.WriteString("\"")
			buffer.WriteString(actualRawMaterialInfo.ActualDate)
			buffer.WriteString("\"")
			
			actualDate = actualRawMaterialInfo.ActualDate
			
			break;
		}
		
		// Check if invoice exists
		f := "getInvoiceAmountById"
		queryArgs := toChaincodeArgs(f, expectedRawMaterialInfo.DemandNumber,expectedRawMaterialInfo.RawMaterialNumber)

		invoiceResponse := stub.InvokeChaincode("a3596e82-9760-494a-bad7-31ffc9530b7e-com-sap-icn-blockchain-invoice-penalty-scenario",queryArgs,"")	

		var invoice Invoice
		json.Unmarshal(invoiceResponse.Payload, &invoice)
		
		buffer.WriteString(", \"invoiceAmount\":")
		buffer.WriteString("\"")
		buffer.WriteString(invoice.InvoiceAmount)
		buffer.WriteString("\"")
		
		timeFormat := "01/02/2006"
		
		if expectedRawMaterialInfo.ExpectedDate != "" && actualDate != "" {
			ed, _ := time.Parse(timeFormat,expectedRawMaterialInfo.ExpectedDate)
			ad, _ := time.Parse(timeFormat,actualDate)
			
			duration := ad.Sub(ed)
			days := (duration.Hours())/24
			
			if days == 0 {
				buffer.WriteString(", \"status\":")
				buffer.WriteString("\"")
				buffer.WriteString("Delivered")
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"state\":")
				buffer.WriteString("\"")
				buffer.WriteString("None")
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"delayPenalty\":")
				buffer.WriteString("\"")
				buffer.WriteString("0.00")
				buffer.WriteString("\"")
			} else {
				buffer.WriteString(", \"status\":")
				buffer.WriteString("\"")
				x := fmt.Sprintf("%.0f",days)
				buffer.WriteString("Delivered+"+x)
				buffer.WriteString("\"")
				
				buffer.WriteString(", \"state\":")
				buffer.WriteString("\"")
				buffer.WriteString("Error")
				buffer.WriteString("\"")
				
				penaltyPercentage := 0
				
				if 	days > 0 && days <= 2 {
					penaltyPercentage = 5
				}else if days >2 && days <= 7{
					penaltyPercentage = 10
				}else if days > 7{
					penaltyPercentage = 20
				}
				
				InvoiceAmountFloat, _ := strconv.ParseFloat((invoice.InvoiceAmount),32)
				InvoicePenalty := (InvoiceAmountFloat*float64(penaltyPercentage)) / 100
				invoicePenaltyStr := fmt.Sprintf("%.2f", InvoicePenalty );
				
				buffer.WriteString(", \"delayPenalty\":")
				buffer.WriteString("\"")
				buffer.WriteString(invoicePenaltyStr)
				buffer.WriteString("\"")
				
			}
		} else {
			buffer.WriteString(", \"delayReason\":")
			buffer.WriteString("\"")
			buffer.WriteString("")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"actualDate\":")
			buffer.WriteString("\"")
			buffer.WriteString("")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"state\":")
			buffer.WriteString("\"")
			buffer.WriteString("Success")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"status\":")
			buffer.WriteString("\"")
			buffer.WriteString("On-Time")
			buffer.WriteString("\"")
			
			buffer.WriteString(", \"delayPenalty\":")
			buffer.WriteString("\"")
			buffer.WriteString("0.00")
			buffer.WriteString("\"")
		}
		
		demandValues, validateErr := stub.GetState(expectedRawMaterialInfo.DemandNumber)
		if validateErr != nil {
			return Error(http.StatusInternalServerError,validateErr.Error())
		}
		
		var demandInfo Demand
		json.Unmarshal(demandValues,&demandInfo)
		
		buffer.WriteString(", \"supplierCode\":")
		buffer.WriteString("\"")
		buffer.WriteString(demandInfo.SupplierCode)
		buffer.WriteString("\"}")
		
		bArrayMemberAlreadyWritten = true
	}
	
	buffer.WriteString("]}")
	return Success(200, "OK", buffer.Bytes())
}