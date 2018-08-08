package main

import (
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"bytes"
	"strings"
)

// key SCHEDULE_{ScheduleId}
// key TRANCHE_(repaymentDate,ScheduleId)
// key TRANCHE-LAST-INDEX value=index
// key ACTIVE_SCHEDULE_{OrderId} <= key of schedule
// key SCHEDULE-LAST-INDEX  value=index

//-----------------------------------------------------------------------------------
// -----------------------------  Schedule story ------------------------------------
//-----------------------------------------------------------------------------------
func (s *SmartContract) createSchedule(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	logger.Info("########### "+projectName+" "+version+" createSchedule ("+args[0]+") ###########")
	// Create schedule for ORDER
	// 0-login,1-OrderId,2-scheduleId,3-amount,4-chargeIssueFee,5-issued,6-activeBefore,7-tranches
	// tranches - array of tranche JSON
	// [ {id,issueDate,repaymentDate,principal,interest,lgot,eachRepaymentFee,rest,scheduleId},
	//    {id,issueDate,repaymentDate,principal,interest,lgot,eachRepaymentFee,rest,scheduleId}
	// ]
	if len(args) != 8 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 8, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	var id uint64 = 0;
	var err error;
	newIdStr := "";
	if(len(args[2])==0) {
		newId, err := s.getCurrentScheduleIndex(stub)
		if err != "" {
			logger.Error(err)
			jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":" + err + "}"
			return shim.Error(jsonResp)
		}
		newId++
		newIdStr = strconv.FormatUint(newId, 10)
		id = newId

		err3 := stub.PutState("SCHEDULE-LAST-INDEX", []byte(newIdStr))
		if err3 != nil {
			logger.Errorf("Can't update SCHEDULE-LAST-INDEX")
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can't update SCHEDULE-LAST-INDEX\"}")
		}
	} else {
		newIdStr = args[2];
		id, err = strconv.ParseUint(args[2], 10, 64);
		if (err != nil) {
			logger.Error("Can not parse ID data")
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not parse ID data (need int64 value)\"}")
		}
	}
	amount,err7 := strconv.ParseFloat(args[3],64);
	if(err7 != nil){
		logger.Error("Can not parse Amount")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Amount (need int64 value)\"}")
	}
	chargeIssueFee, err4:= strconv.ParseBool(args[4])
	if(err4 != nil){
		logger.Error("Can not parse chargeIssueFee")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse chargeIssueFee (need bool value)\"}")
	}
	issued, err8:= strconv.ParseBool(args[5])
	if(err8 != nil){
		logger.Error("Can not parse issued")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse issued (need bool value)\"}")
	}

	schedKey,err2 := stub.CreateCompositeKey("SCHEDULE_", []string{newIdStr})
	if(err2!=nil){
		logger.Error("Can not serializated Schedule")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not serializated Schedule\"}")
	}
	order,orderKeyId,err3 := s.getOrderByIdInner(stub,args[1])
	if err3 != "" {
		logger.Error(err3)
		return shim.Error(err3)
	}
	order.ScheduleIds = s.appendData(order.ScheduleIds, newIdStr)

	trancheIds := []string{};
	var tranche []Tranche;
	err = json.Unmarshal([]byte(args[7]), &tranche)
	if err != nil {
		logger.Error("Can not parse Tranche")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Tranche (need JSON array)"+err.Error()+"\"}")
	}
	for i :=range tranche {
		tranche[i].ScheduleId = newIdStr;

		trancheAsBytes, _ := json.Marshal(tranche[i])
		compositeKey:= "TRANCHE_"+s.createKeyDate(tranche[i].RepaymentDate)+newIdStr;
		err1:= stub.PutState(compositeKey, trancheAsBytes)

		if(err1 != nil){
			logger.Error("Can not add Tranche="+ string(trancheAsBytes))
		} else {
			logger.Info("Added: " + string(trancheAsBytes))
			trancheIds = s.appendData(trancheIds, compositeKey)
		}
	}

	schedule := &Schedule{Id: id,
		CreationDate: s.timeStamp(),
		Amount: amount,
		ChargeIssueFee: chargeIssueFee,
		Issued: issued,
		ActiveBefore: args[6],
		Tranches: trancheIds,
		OrderId: args[1]}

	scheduleAsBytes, _ := json.Marshal(schedule)
	err1:= stub.PutState(schedKey, scheduleAsBytes)
	if(err1 != nil){
		logger.Error("Can not add Schedule")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not add Schedule\"}")
	}
	orderAsBytes,_ := json.Marshal(order)
	err1= stub.PutState(orderKeyId, orderAsBytes)
	if(err1 != nil){
		logger.Error("Can not change Order")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not change Order\"}")
	}
	logger.Info("########### "+projectName+" "+version+" success createSchedule ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"scheduleId\":"+newIdStr+",\"description\":\"Schedule is successfully added\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) setActiveSchedule(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - ScheduleId,
	// 2 -OrderId

	logger.Info("########### "+projectName+" "+version+" setActiveSchedule ("+args[1]+") ###########")
	if len(args) != 3 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	schedKey,err2 := stub.CreateCompositeKey("SCHEDULE_", []string{args[1]})
	if(err2!=nil){
		logger.Error("Can not create Schedule composite key")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not create Schedule composite key\"}")
	}
	schedAsBytes, err := stub.GetState(schedKey)
	if(err!=nil){
		logger.Error("Can not get Schedule")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not get Schedule\"}")
	}
	if(len(schedAsBytes)==0){
		logger.Error("Can not get Schedule, not found")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not get Schedule, not found\"}")
	}
	activeKey,err3 := stub.CreateCompositeKey("ACTIVE_SCHEDULE_", []string{args[2]})
	if(err3!=nil){
		logger.Error("Can not create Schedule active composite key")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not create Schedule active composite key\"}")
	}
	err1:= stub.PutState(activeKey, []byte(schedKey))
	if(err1 != nil){
		logger.Error("Can not add Schedule")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not set active Schedule\"}")
	}

	logger.Info("########### "+projectName+" "+version+" success setActiveSchedule ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Schedule is successfully set Actived\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getActiveSchedule(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - OrderId

	logger.Info("########### " + projectName + " " + version + " getActiveSchedule (" + args[0] + ") ###########")
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}

	activeKey,err3 := stub.CreateCompositeKey("ACTIVE_SCHEDULE_", []string{args[1]})
	if(err3!=nil){
		logger.Error("Can not create Schedule active composite key")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not create Schedule active composite key\"}")
	}
	schedKey, err1:= stub.GetState(activeKey)
	if(err1 != nil){
		logger.Error("Can not get Active Schedule")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not get active Schedule\"}")
	}
	result, err := s.getScheduleByIdInner(stub, "",string(schedKey), args)
	if(err != ""){
		return shim.Error(err);
	}
	logger.Info("########### "+projectName+" "+version+" success getActiveSchedule ("+args[0]+") ###########")
	return shim.Success(result);
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getScheduleById(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - ScheduleId

	logger.Info("########### " + projectName + " " + version + " getScheduleById (" + args[0] + ") ###########")
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	result, err := s.getScheduleByIdInner(stub, args[1],"", args)
	if(err != ""){
		return shim.Error(err);
	}
	logger.Info("########### "+projectName+" "+version+" success getScheduleById ("+args[0]+") ###########")
	return shim.Success(result);
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getScheduleByIdInner(stub shim.ChaincodeStubInterface, id string, key string, args []string) ([]byte, string) {
	var buffer bytes.Buffer
	var schedKey string
	var err2 error

	if(len(key) > 0){
		schedKey = key
	} else {
		schedKey,err2 = stub.CreateCompositeKey("SCHEDULE_", []string{id})
		if(err2!=nil){
			logger.Error("Can not create key for Schedule")
			return nil, "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not create key for Schedule\"}"
		}
	}

	schedAsBytes, err2:= stub.GetState(schedKey)
	if(err2 != nil){
		logger.Error("Can not get Schedule")
		return nil, "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not get Schedule\"}"
	}
	if(len(schedAsBytes)== 0){
		logger.Error("Schedule not found")
		return nil, "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Schedule not found\"}"
	}
	schedule := Schedule{}
	err := json.Unmarshal(schedAsBytes, &schedule)
	if err != nil {
		logger.Error("Can not parse Schedule")
		return nil, "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Schedule\"}"
	}
	chargeIssue := "";
	if(schedule.ChargeIssueFee == true){
		chargeIssue ="true"
	} else {
		chargeIssue ="false"
	};
	issued := "";
	if(schedule.Issued == true){
		issued ="true"
	} else {
		issued ="false"
	};

	buffer.WriteString("{\"id\":"+strconv.FormatUint(schedule.Id, 10)+
		",\"creationDate\":"+strconv.FormatInt(schedule.CreationDate,10)+
		",\"amount\":"+strconv.FormatFloat(schedule.Amount, 'E', -1, 64)+
		",\"chargeIssueFee\":"+ chargeIssue+
		",\"issued\":"+issued+
		",\"orderId\":"+schedule.OrderId+
		",\"activeBefore\":"+schedule.ActiveBefore+
		",\"tranches\":[");

	alreadyWritten := false;
	for i := range schedule.Tranches {
		if(len(schedule.Tranches[i])>0){
			trancheAsBytes, err2:= stub.GetState(schedule.Tranches[i])
			if(err2 != nil){
				logger.Error("Can not get Tranche")
				return nil, "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not get Tranche\"}"
			}
			if (len(trancheAsBytes)>0) {
				if alreadyWritten == true {
					buffer.WriteString(",")
					alreadyWritten = false
				}
				buffer.WriteString(string(trancheAsBytes))
				alreadyWritten = true
			}
		}
	}

	buffer.WriteString("]}")
	return buffer.Bytes(), ""
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getScheduleList(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - OrderId
	logger.Info("########### " + projectName + " " + version + " getScheduleList (" + args[0] + ") ###########")
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	order,_,err2 := s.getOrderByIdInner(stub,args[1])
	if err2 != "" {
		logger.Error(err2)
		return shim.Error(err2)
	}
	var buffer bytes.Buffer
	buffer.WriteString("[");
	alreadyWritten := false;

	for i := range order.ScheduleIds {
		result, err := s.getScheduleByIdInner(stub,order.ScheduleIds[i],"",args)
		if(err != ""){
			return shim.Error(err);
		}
		if(len(result)>0) {
			if alreadyWritten == true {
				buffer.WriteString(",")
				alreadyWritten = false
			}
			buffer.WriteString(string(result))
			alreadyWritten = true
		}
	}
	buffer.WriteString("]")

	logger.Info("########### "+projectName+" "+version+" success getScheduleList ("+args[0]+") ###########")
	return shim.Success(buffer.Bytes());
}
//------------------------------------------------------------------------------------
func (s *SmartContract) createKeyDate(date string) string {
	// format '2017-06-02' => '20170602'
	val := strings.Split(date,"-")
	if(len(val)==3) {
		return val[0] + val[1] + val[2]
	} else {
		return date
	};
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getListOfTranchesByDate(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - dateStart  format '2017-06-02'
	// 2 - dateEnd    format '2017-06-02'
	scMap := make(map[string]Schedule)
	ordMap := make(map[string]Order)

	logger.Info("########### " + projectName + " " + version + " getListOfTranchesByDate (" + args[0] + ") ###########")
	if len(args) != 3 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	startKey:= "TRANCHE_"+s.createKeyDate(args[1]);
	finishKey:= "TRANCHE_"+s.createKeyDate(args[2]);

	resultsIterator, err := stub.GetStateByRange(startKey, finishKey)
	if err != nil {
		logger.Errorf("Can't get by range data")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get by range data\"}"
		return shim.Error(jsonResp)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("{\"start\":\""+args[1]+"\",\"finish\":\""+args[2]+"\",\"result\":[")
	alreadyWritten := false;

	for resultsIterator.HasNext() {
		queryResponseBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Errorf("Can't get next value by range data")
			jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get next value by range data\"}"
			return shim.Error(jsonResp)
		}
		queryResponseValue := queryResponseBytes.Value
		queryResponse := string(queryResponseValue)

		if(len(queryResponse) > 0) {
			tranche:= Tranche{}
			err = json.Unmarshal([]byte(queryResponse), &tranche)
			if err != nil {
				logger.Error("Can not parse Tranche")
				return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Tranche - "+err.Error()+"\"}")
			}

			if tranche.EachRepaymentFee < (tranche.Interest+tranche.Principal) {
				// get Schedule
				schedule, prs := scMap[tranche.ScheduleId]
				schedKey, err2 := stub.CreateCompositeKey("SCHEDULE_", []string{tranche.ScheduleId})
				if (err2 != nil) {
					logger.Error("Can not get Schedule key")
					return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not get Schedule key\"}")
				}

				if prs == false {
					// get schedule by tranche
					schedAsBytes, err := stub.GetState(string(schedKey))
					if (err != nil) {
						logger.Error("Can not get Schedule")
						return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not get Schedule\"}")
					}
					err = json.Unmarshal(schedAsBytes, &schedule)
					if err != nil {
						logger.Error("Can not parse Schedule")
						return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not parse Schedule - " + err.Error() + "(" + string(schedAsBytes) + ")\"}")
					}
					scMap[tranche.ScheduleId] = schedule;
				}

				// get Order
				order, prs := ordMap[schedule.OrderId]
				err5 := "";

				if prs == false {
					order, _, err5 = s.getOrderByIdInner(stub, schedule.OrderId);
					if len(err5) == 0 {
						ordMap[schedule.OrderId] = order;
					}
				}

				if len(err5) == 0 && order.Status == "ACTIVE" {

					// get OrderId, get data active schedule or not
					activeKey, err := stub.CreateCompositeKey("ACTIVE_SCHEDULE_", []string{schedule.OrderId})
					if (err != nil) {
						logger.Error("Can not create Schedule active composite key")
						return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not create Schedule active composite key\"}")
					}

					value, _ := stub.GetState(activeKey)
					// if schedule is active=true  => adding to result
					if (string(value) == schedKey) {
						if alreadyWritten == true {
							buffer.WriteString(",")
							alreadyWritten = false;
						}
						buffer.WriteString(queryResponse)
						alreadyWritten = true
					}
				}
			}
		}
	}
	//buffer.WriteString("],\"description\":"+desc+"]}")
	buffer.WriteString("]}")
	logger.Info("########### "+projectName+" "+version+" success getListOfTranchesByDate ("+args[0]+") ###########")
	return shim.Success(buffer.Bytes());
}
//------------------------------------------------------------------------------------
func (s *SmartContract) paymentTranche(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - Login,
	// 1 - dateOfTranche  format '2017-06-02'
	// 2 - scheduleId
	// 3 - eachRepaymentFee

	logger.Info("########### " + projectName + " " + version + " paymentTranche (" + args[0] + ") ###########")
	if len(args) != 4 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 4, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	eachRepaymentFee,err2 := strconv.ParseFloat(args[3],10);
	if(err2 != nil){
		logger.Error("Can not parse eachRepaymentFee data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse eachRepaymentFee data (need int64 value)\"}")
	}

	compositeKey:= "TRANCHE_"+s.createKeyDate(args[1])+args[2];
	tranche := Tranche{};
	trancheAsBytes, err2 := stub.GetState(compositeKey)
	if (err2 != nil) {
		logger.Error("Tranche not found")
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Tranche not found\"}")
	}
	if len(trancheAsBytes) == 0 {
		logger.Error("Could not found Tranche")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Could not found Tranche\"}")
	}
	err2 = json.Unmarshal(trancheAsBytes, &tranche)
	if err2 != nil {
		logger.Error("Can not parse Tranche")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Tranche - "+err2.Error()+"("+string(trancheAsBytes)+")\"}")
	}

	if(tranche.Id == 0) {
		newId, err := s.getCurrentTrancheIndex(stub)
		if err != "" {
			logger.Error(err)
			jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":" +err+"}"
			return shim.Error(jsonResp)
		}
		newId++
		newIdStr := strconv.FormatUint(newId, 10)
		tranche.Id = newId

		err3 := stub.PutState("TRANCHE-LAST-INDEX", []byte(newIdStr))
		if err3 != nil {
			logger.Errorf("Can't update TRANCHE-LAST-INDEX")
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update TRANCHE-LAST-INDEX\"}")
		}
	}
	tranche.EachRepaymentFee = eachRepaymentFee

	trancheAsBytes, _ = json.Marshal(tranche)
	err2 = stub.PutState(compositeKey, trancheAsBytes)
	if(err2 != nil){
		logger.Error("Can not change Tranche")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not change Tranche\"}")
	}

	logger.Info("########### "+projectName+" "+version+" success paymentTranche ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Tranche is successfully changed\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getCurrentScheduleIndex(stub shim.ChaincodeStubInterface) (uint64, string) {
	ind, err := stub.GetState("SCHEDULE-LAST-INDEX")
	if err != nil {
		return 0, "\"Failed to get SCHEDULE-LAST-INDEX: "+err.Error()+"\""
	} else {
		if len(ind)>0 {
			index, err := strconv.ParseUint(string(ind),10,64)
			if err != nil {
				return 0, "\"Can't create sequence SCHEDULE-LAST-INDEX\""
			} else {
				return index, ""
			}
		} else {
			return 0, "\"Can't get sequence SCHEDULE-LAST-INDEX=nil\""
		}
	}
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getCurrentTrancheIndex(stub shim.ChaincodeStubInterface) (uint64, string) {
	ind, err := stub.GetState("TRANCHE-LAST-INDEX")
	if err != nil {
		return 0, "\"Failed to get TRANCHE-LAST-INDEX: "+err.Error()+"\""
	} else {
		if len(ind)>0 {
			index, err := strconv.ParseUint(string(ind),10,64)
			if err != nil {
				return 0, "\"Can't create sequence TRANCHE-LAST-INDEX\""
			} else {
				return index, ""
			}
		} else {
			return 0, "\"Can't get sequence TRANCHE-LAST-INDEX=nil\""
		}
	}
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getUserTranches(stub shim.ChaincodeStubInterface, user User, args []string) pb.Response {
	// 0 - Login,
	// 1 - dateStart format '2017-06-02'
	// 2 - dateFinish format '2017-06-02'

	logger.Info("########### " + projectName + " " + version + " getUserTranches (" + args[0] + ") ###########")
	if len(args) != 3 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	var buffer bytes.Buffer
	buffer.WriteString("{\"start\":\"" + args[1] + "\",\"finish\":\"" + args[2] + "\",\"result\":[")
	alreadyWritten := false;
	order := Order{}

	for i := range user.Orders {
		orderStr, err := stub.GetState(user.Orders[i])
		if err != nil {
			logger.Error("Can not get Order")
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not get Order\"}")
		}
		err = json.Unmarshal(orderStr, &order)
		if err != nil {
			logger.Error("Can not unmarshal Order")
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not unmarshal Order\"}")
		}

		if order.Status == "ACTIVE" {
			activeKey, err3 := stub.CreateCompositeKey("ACTIVE_SCHEDULE_", []string{order.Id})
			if (err3 != nil) {
				logger.Error("Can not create Schedule active composite key")
				return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not create Schedule active composite key\"}")
			}
			schKey, err1 := stub.GetState(activeKey)
			if (err1 != nil) {
				logger.Error("Can not get Schedule key")
				return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not get active Schedule key\"}")
			}
			scheduleAsBytes, err1 := stub.GetState(string(schKey))
			if (err1 != nil) {
				logger.Error("Can not get Schedule key2")
				return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not get active Schedule key2\"}")
			}
			if len(scheduleAsBytes)>0 {
				schedule := Schedule{};
				err := json.Unmarshal(scheduleAsBytes, &schedule)
				if err != nil {
					logger.Error("Can not parse Schedule")
					return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not parse Schedule - " + err.Error() + "(" + string(scheduleAsBytes) + ")" + "\"}")
				}

				schStr := strconv.FormatUint(schedule.Id, 10)
				startKey := "TRANCHE_" + s.createKeyDate(args[1]) + schStr
				finishKey := "TRANCHE_" + s.createKeyDate(args[2]) + schStr

				resultsIterator, err := stub.GetStateByRange(startKey, finishKey)
				if err != nil {
					logger.Errorf("Can't get by range data")
					jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can't get by range data\"}"
					return shim.Error(jsonResp)
				}
				defer resultsIterator.Close()

				for resultsIterator.HasNext() {
					queryResponseBytes, err := resultsIterator.Next()
					if err != nil {
						logger.Errorf("Can't get next value by range data")
						jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can't get next value by range data\"}"
						return shim.Error(jsonResp)
					}
					queryResponseValue := queryResponseBytes.Value
					queryResponse := string(queryResponseValue)

					if (len(queryResponse) > 0) {
						if alreadyWritten == true {
							buffer.WriteString(",")
							alreadyWritten = false;
						}
						buffer.WriteString(queryResponse)
						alreadyWritten = true
					}
				}
			}
		}
	}
	buffer.WriteString("]}")
	logger.Info("########### "+projectName+" "+version+" success getUserTranches ("+args[0]+") ###########")
	return shim.Success(buffer.Bytes());
}
