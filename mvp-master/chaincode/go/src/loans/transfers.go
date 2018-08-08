package main

import (
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"bytes"
)

// key TRANS_{userEth, date}
// key TRANSFER_(date)

//-----------------------------------------------------------------------------------
// -----------------------------  Transfer story ------------------------------------
//-----------------------------------------------------------------------------------

func (s *SmartContract) tokenTransfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, trHash-1, from - 2, to - 3, amount - 4, 5 - orderId}
	logger.Info("########### " + projectName + " " + version + " tokenTransfer (" + args[0] + ") ###########")
	var types string
	if len(args) != 6 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 6, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[1] == "" || args[2] =="" {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null Tx and Rx data\"}")
	}
	if args[3] =="" {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null amount data\"}")
	}
	amount, err7 := strconv.ParseFloat(args[4], 64)
	if(err7 != nil){
		logger.Error("Can not parse amount data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse amount data (need float value)\"}")
	}
	trans := &Transfer{Tx: args[1],
		From: args[2],
		To: args[3],
		Amount: amount,
		OrderId: args[4]}

	tr, _ := json.Marshal(trans)
	logger.Info("update data Payments = "+string(tr))

	// TRANSFER
	compositeTransfer := "TRANSFER_"+s.timeStampStr()
	logger.Info("compositeTransfer = ", compositeTransfer)

	err3 := stub.PutState(compositeTransfer, tr)
	if err3 != nil {
		logger.Errorf("Can't add PAYMENT KEY value")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't add PAYMENT KEY value\"}")
	}
	// FROM
	compositeFrom := "TRANS_"+args[2]+s.timeStampStr()
	logger.Info("compositeFrom = ", compositeFrom)

	err4 := stub.PutState(compositeFrom, []byte(compositeTransfer))
	if err4 != nil {
		logger.Errorf("Can't add FROM KEY value")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't add FROM KEY value\"}")
	}
	// Update balance
	userFrom,fromKey,err11:= s.getUserByEthAddress(stub, args[2], args)
	if err11 != "" {
		return shim.Error(err11)
	}
	err11 = s.changeUserBalanceInner(stub, userFrom, string(fromKey), []string{args[2],args[2],"-"+args[4]})
	if err11 != "" {
		return shim.Error(err11)
	}
	// TO
	compositeTo := "TRANS_"+args[3]+s.timeStampStr()
	logger.Info("compositeTo = ", compositeTo)

	err6 := stub.PutState(compositeTo, []byte(compositeTransfer))
	if err6 != nil {
		logger.Errorf("Can't add TO KEY value")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't add TO KEY value\"}")
	}
	// Update balance
	userTo,toKey,err11:= s.getUserByEthAddress(stub, args[3], args)
	if err11 != "" {
		return shim.Error(err11)
	}
	err11 = s.changeUserBalanceInner(stub, userTo, string(toKey), []string{args[3],args[3],args[4]})
	if err11 != "" {
		return shim.Error(err11)
	}

	logger.Info("########### "+projectName+" "+version+" success tokenTransfer ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order for "+types+" added successfully\"}"))
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getHistoryTransferToken(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	var startKey string;
	var finishKey string;

	// 0 - login
	// 1 - eth
	// 2 - periodStart
	// 3 - periodEnd
	if len(args) != 4 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 4, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	logger.Info("########### "+projectName+" "+version+" getHistoryTransferToken ###########")
	var periodStart string;
	var periodFinish string;
	if(len(args[2])== 0){
		periodStart = "0"
	} else {
		periodStart = args[2]
	};
	if(len(args[3])== 0){
		periodFinish = s.timeStampStr()  //time.Now().String()
	} else {
		periodFinish = args[3]
	};

	if(len(args[1]) == 0){
		startKey = "TRANSFER_"+periodStart;
		finishKey = "TRANSFER_"+periodFinish;

	} else {
		startKey = "TRANS_"+args[1]+periodStart;
		finishKey = "TRANS_"+args[1]+periodFinish;
	}

	resultsIterator, err := stub.GetStateByRange(startKey, finishKey)
	if err != nil {
		logger.Errorf("Can't get by range data")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"["+startKey+","+finishKey+"]Can't get by range data ("+err.Error()+")\"}"
		return shim.Error(jsonResp)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("{\"start\":\""+periodStart+"\",\"finish\":\""+periodFinish+"\",\"result\":[")

	alreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponseBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Errorf("Can't get next value by range data")
			jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get next value by range data\"}"
			return shim.Error(jsonResp)
		}
		queryResponseValue := queryResponseBytes.Value
		queryResponse := string(queryResponseValue)

		if(len(queryResponseValue)>0) {
			if alreadyWritten == true {
				buffer.WriteString(",")
			}
			alreadyWritten = false

			if (len(args[1]) > 0) {
				value, err11 := stub.GetState(string(queryResponseValue))
				if err11 != nil {
					logger.Error("Failed to get TRANSFER key value")
					jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to get TRANSFER key value\"}"
					return shim.Error(jsonResp)
				} else {
					queryResponseValue = value
				}
			}
			payment := Transfer{}
			err = json.Unmarshal(queryResponseValue, &payment)
			logger.Info("payment = ", payment)
			if err != nil {
				logger.Error("[" + string(queryResponse) + "] Failed to decode JSON")
				return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"[" + string(queryResponse) + "] Failed to decode JSON\"}")
			}
			buffer.WriteString(queryResponse)
			alreadyWritten = true

		}
	}
	buffer.WriteString("]}")

	logger.Info("########### "+projectName+" "+version+" success getHistoryTransferToken ("+args[0]+") ###########")
	return shim.Success(buffer.Bytes());
}
