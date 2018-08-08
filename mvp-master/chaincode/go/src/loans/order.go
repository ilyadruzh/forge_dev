package main

import (
	"encoding/json"
	"strconv"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"bytes"
)

// key ORDER-LAST-INDEX value=index
// key ORDER-ID{orderId}
// key ORDER-ACTIVE value=index
// key ACTIVE_OFFERS_BORROWS
// key ACTIVE_OFFERS_LENDS

//-----------------------------------------------------------------------------------
// -----------------------------  Order story ---------------------------------------
//-----------------------------------------------------------------------------------
func (s *SmartContract) getCurrentOrderIndex(stub shim.ChaincodeStubInterface) (int, string) {
	ind, err := stub.GetState("ORDER-LAST-INDEX")
	if err != nil {
		return -1, "\"Failed to get ORDER-LAST-INDEX: "+err.Error()+"\""
	} else {
		index := 0
		if ind != nil {
			index, err = strconv.Atoi(string(ind))
			if err != nil {
				return -1, "\"Can't create sequence ORDER-LAST-INDEX\""
			} else {
				return index, ""
			}
		} else {
			return -1, "\"Can't get sequence ORDER-LAST-INDEX=nil\""
		}
	}
}

func (s *SmartContract) createOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, period, Borrower, Lender, Sum, Percent, StartDate, FinishDate}
	var types string
	if len(args) != 8 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 8, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[2] == "" && args[3] =="" {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null BORROWER='"+args[2]+"' and LENDER='"+args[3]+"'\"}")
	}
	percent,err5 := strconv.ParseFloat(args[5], 32)
	if(err5 != nil){
		logger.Error("Can not parse Percent data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Percent data (need float32 value)\"}")
	}
	period,err3 := strconv.Atoi(args[1]);
	if(err3 != nil){
		logger.Error("Can not parse period data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Period data (need int value)\"}")
	}
	sum, err6 := strconv.ParseFloat(args[4], 64)
	if(err6 != nil){
		logger.Error("Can not parse Sum data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Sum data (need float value)\"}")
	}

	newId, err := s.getCurrentOrderIndex(stub)
	if err != "" {
		logger.Error(err)
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":" +err+ "}"
		return shim.Error(jsonResp)
	}
	newId++
	newIdStr := strconv.Itoa(newId)

	order := &Order{Id: newIdStr,
		Period: period,
		Borrower: args[2],
		Lender: args[3],
		Sum: sum,
		Percent: percent,
		CreateDate: time.Now().String(),
		StartDate: args[6],
		FinishDate: args[7],
		ContractId: 0,
		Status: "NEW",
		ScheduleIds: []string{}}

	compositeKeyId, err2 := stub.CreateCompositeKey("ORDER-ID", []string{newIdStr})
	logger.Info("compositeKey = ", compositeKeyId)
	if err2 != nil {
		logger.Errorf("Can't create siquence ORDER-ID for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create siquence ORDER-ID\"}")
	}
	ord, _ := json.Marshal(order)
	logger.Info("return data orderJson = "+string(ord))
	err3 = stub.PutState(compositeKeyId, ord)
	if err3 != nil {
		logger.Errorf("Can't put data ORDER for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't put data ORDER\"}")
	}
	err3 = stub.PutState("ORDER-LAST-INDEX", []byte(newIdStr))
	if err3 != nil {
		logger.Errorf("Can't update ORDER-LAST-INDEX")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update ORDER-LAST-INDEX\"}")
	}

	if args[2]!= "" {
		if args[3]!= ""{
			types = "BOTH"
		} else {
			types = "BORROWER"
		}
		err4 := s.addOrderToUser(stub, "BORROWER", args[2], compositeKeyId)
		if err4 != "" {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err4+"\"}")
		}

	} else {
		types = "LENDER"
	}
	if args[3]!="" {
		err4 :=s.addOrderToUser(stub, "LENDER", args[3], compositeKeyId)
		if err4 != "" {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err4+"\"}")
		}
	}
	//compositeKeyId2, err2 := stub.CreateCompositeKey("OFFER-ACTIVE", []string{args[1]})
	//logger.Info("compositeKey = ", compositeKeyId2)
	//if err2 != nil {
	//	logger.Error("Can't get siquence OFFER-ACTIVE for login="+args[0])
	//	return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get siquence OFFER-ACTIVE\"}")
	//}
	//err2 = stub.PutState(compositeKeyId2, []byte(""))
	//if err2 != nil {
	//	logger.Error("Failed to get siquence OFFER-ACTIVE ("+args[0]+"): "+err2.Error())
	//	jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to get siquence OFFER-ACTIVE "+err2.Error()+"\"}"
	//	return shim.Error(jsonResp)
	//}

	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"orderId\":"+newIdStr+",\"status\":true,\"description\":\"Order for "+types+" added successfully\"}"))
}

func (s *SmartContract) addOrderToUser(stub shim.ChaincodeStubInterface, keyName string, login string, data string) string {
	// addOrderToUser(stub, "BORROWER", args[4], compositeKeyId)

	user, compositeKey, err := s.getUserInner(stub, []string{login})
	if err != "" {
		return string(err)
	}
	user.Orders = s.appendData(user.Orders, data)

	usr, _ := json.Marshal(user)
	err2 := stub.PutState(compositeKey, usr)

	if err2 != nil {
		logger.Error("Failed add ORDER to USER ("+login+"): "+err2.Error())
		return "Failed add ORDER to USER ("+login+"): "+err2.Error()
	}
	return ""
}

func (s *SmartContract) getOrderByIdInner(stub shim.ChaincodeStubInterface, id string) (Order, string, string) {
	var order Order
	compositeKeyId, err2 := stub.CreateCompositeKey("ORDER-ID", []string{id})
	logger.Info("getOrderByIdInner compositeKey = ", compositeKeyId)
	if err2 != nil {
		return order,compositeKeyId, "Can't create siquence ORDER-ID for id="+id
	}
	orderStr, err := stub.GetState(compositeKeyId)
	if err != nil {
		return order,compositeKeyId, "\"Failed to get ORDER by id="+id
	} else {
		err := json.Unmarshal(orderStr, &order)
		if err != nil {
			return order,compositeKeyId, "\"Failed to get ORDER by id=" + id
		} else {
			return order,compositeKeyId,""
		}
	}
}

func (s *SmartContract) activateOffer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if args[1] == "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not be null orderId=''\"}")
	}
	var msg string
	order, orderKey, err := s.getOrderByIdInner(stub, args[1])
	if err != "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"" + err + "\"}")
	}
	if order.Status == "NEW" {
		msg = "Changed status from 'NEW' to 'ACTIVE_OFFER'"
		order.Status = "ACTIVE_OFFER"
		key := "";
		if(order.Borrower==""){key = "ACTIVE_OFFERS_LENDS"
			} else {key = "ACTIVE_OFFERS_BORROWS"}

		arr, err5 := s.getBorrowOffersIdsList(stub, key, args);
		if(err5 != ""){
			return shim.Error(err5)
		}
		arr = s.appendData(arr, args[1])
		offers,_ := json.Marshal(arr)
		err2 := stub.PutState(key, offers)
		if err2 != nil {
			logger.Errorf("Can't update ACTIVE_OFFERS array("+string(offers)+")")
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update ACTIVE_OFFERS array("+string(offers)+")}")
		}
		msg = "Status changed '"+order.Status+"' to 'ACTIVE_OFFER' successful"
	} else {
		msg = "You can't change status '"+order.Status+"' to 'ACTIVE_OFFER'"
		logger.Error(msg)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+msg+"\"}")
	}

	ord, _ := json.Marshal(order)
	logger.Info("return data orderJson = "+string(ord))
	err2 := stub.PutState(orderKey, ord)
	if err2 != nil {
		logger.Errorf("Can't update status ORDER for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update status ORDER\"}")
	}
	logger.Info(msg)
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\""+msg+"\"}"))
}

func (s *SmartContract) confirmOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[1] == ""  {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null orderId=''\"}")
	}

	var msg string
	order, orderKey, err := s.getOrderByIdInner(stub, args[1])
	if err != "" {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err+"\"}")
	}
	if(args[0] == order.Borrower){
		if order.Status == "NEW" || order.Status == "ACTIVE_OFFER" {
			if order.Status == "ACTIVE_OFFER"{
				// remove from ACTIVE_OFFER array
				key := ""
				if(len(order.Borrower)==0){key = "ACTIVE_OFFERS_LENDS"
				} else {key = "ACTIVE_OFFERS_BORROWS"}
				arr, err5 := s.getBorrowOffersIdsList(stub, key, args);
				if(err5 != ""){
					return shim.Error(err5)
				}
				arr = s.delElem(arr,args[1]);
				offers,_ := json.Marshal(arr)
				err2 := stub.PutState(key, offers)
				if err2 != nil {
					logger.Errorf("Can't update ACTIVE_OFFERS array("+string(offers)+")")
					return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update ACTIVE_OFFERS array("+string(offers)+")}")
				}
			}
			msg = "Changed status from 'NEW' to 'CONFIRMED_BY_BORROWER'"
			order.Status = "CONFIRMED_BY_BORROWER"
		} else
		if order.Status == "CONFIRMED_BY_BORROWER" || order.Status == "CONFIRMED" {
			// return already confirmed
			msg = "Order already confirmed, current status='"+order.Status+"'"
			return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already confirmed, current status='"+order.Status+"'\"}"))
		} else
		if order.Status == "CONFIRMED_BY_LENDER" {
			msg ="Change status from 'CONFIRMED_BY_LENDER' to 'CONFIRMED'"
			order.Status = "CONFIRMED"
		} else {
			msg ="Can't change status from '"+order.Status+"'"
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+msg+"\"}")
		}
	} else
	if(args[0] == order.Lender || len(order.Lender)==0){
		if(len(order.Lender)==0){
			order.Lender = args[0];
		}
		if order.Status == "NEW" || order.Status == "ACTIVE_OFFER" {
			if order.Status == "ACTIVE_OFFER"{
				// remove from ACTIVE_OFFER array
				key := ""
				if(len(order.Borrower)==0){key = "ACTIVE_OFFERS_LENDS"
				} else {key = "ACTIVE_OFFERS_BORROWS"}
				arr, err5 := s.getBorrowOffersIdsList(stub, key, args);
				if(err5 != ""){
					return shim.Error(err5)
				}
				arr = s.delElem(arr,args[1]);
				offers,_ := json.Marshal(arr)
				err2 := stub.PutState(key, offers)
				if err2 != nil {
					logger.Errorf("Can't update ACTIVE_OFFERS array("+string(offers)+")")
					return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update ACTIVE_OFFERS array("+string(offers)+")}")
				}
			}
			msg = "Changed status from 'NEW' to 'CONFIRMED_BY_LENDER'"
			order.Status = "CONFIRMED_BY_LENDER"
		} else
		if order.Status == "CONFIRMED_BY_LENDER" || order.Status == "CONFIRMED" {
			// return already confirmed
			msg = "Order already confirmed, current status='"+order.Status+"'"
			return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already confirmed, current status='"+order.Status+"'\"}"))
		} else
		if order.Status == "CONFIRMED_BY_BORROWER" {
			msg = "Change status from 'CONFIRMED_BY_BORROWER' to 'CONFIRMED'"
			order.Status = "CONFIRMED"
		} else {
			msg ="Can't change status from '"+order.Status+"'"
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+msg+"\"}")
		}
	} else {
		msg = "You can't change status, only BORROWER or LENDER should do it"
		logger.Error(msg)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+msg+"\"}")
	}

	ord, _ := json.Marshal(order)
	logger.Info("return data orderJson = "+string(ord))
	err2 := stub.PutState(orderKey, ord)
	if err2 != nil {
		logger.Errorf("Can't update status ORDER for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't update status ORDER\"}")
	}
	logger.Info(msg)
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\""+msg+"\"}"))
}

func (s *SmartContract) getLends(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, LenderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[1] == ""  {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null LenderId=''\"}")
	}

	res, err := s.getOrders(stub, args, "LENDER")
	if err != "" {
		if err == "Empty list" {
			err = "{\"login\":\""+args[0]+"\", \"status\":true,\"description\":\"Empty list\"}"
			return shim.Success([]byte(err));
		} else {
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"" + err + "\"}")
		}
	}
	offr, _ := json.Marshal(res)
	logger.Info("return data ordersJson = "+string(offr))
	logger.Info("########### "+projectName+" "+version+" success getLendOrders for user="+args[0]+" ###########")
	return shim.Success(offr);
}

func (s *SmartContract) getBorrows(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, BorrowerId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[1] == ""  {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null BorrowerId=''\"}")
	}

	res, err := s.getOrders(stub, args, "BORROWER")
	if err != "" {
		if err == "Empty list" {
			err = "{\"login\":\""+args[0]+"\", \"status\":true,\"description\":\"Empty list\"}"
			return shim.Success([]byte(err));
		} else {
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"" + err + "\"}")
		}
	}
	offr, _ := json.Marshal(res)
	logger.Info("return data ordersJson = "+string(offr))
	logger.Info("########### "+projectName+" "+version+" success getBorrowOrders for user="+args[0]+" ###########")
	return shim.Success(offr);
}

func (s *SmartContract) getOrders(stub shim.ChaincodeStubInterface, args []string, key string) ([]Order, string) {
	var res []Order
	orders, err := s.getAllOrders(stub, args, key);

	if err=="" {
		for i:=0; i<len(orders); i++ {
			if key == "BORROWER" &&	orders[i].Borrower == args[0] {
				res = append(res, orders[i])
			} else
			if key == "LENDER" &&	orders[i].Lender == args[0] {
				res = append(res, orders[i])
			}
		}
		if len(res)==0 {
			return res, "Empty list"
		} else {
			return res, "";
		}
	} else {
		return orders, err
	}
}

func (s *SmartContract) getAllOrders(stub shim.ChaincodeStubInterface, args []string, key string) ([]Order, string) {

	var res []Order
	var order Order

	if len(args) != 2 {
		return res, "Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args))
	}
	if args[1] == "" {
		return res, "Can not be null " + key + "='" + args[1] + "'"
	}
	user, _, err := s.getUserInner(stub, []string{args[1]})
	if err != "" {
		return res, string(err)
	}
	if len(user.Orders)>0 {
		for i:=0; i < len(user.Orders); i++ {
			data, err := stub.GetState(user.Orders[i])

			if err == nil {
				err = json.Unmarshal([]byte(data), &order)
				logger.Info("order = ", user)
				if err != nil {
					logger.Error("Failed to decode JSON of: " +user.Orders[i]+", value=["+string(data)+"]")
				} else {
					res = append(res, order)
				}
			}
		}
	} else {
		return res, "Empty list"
	}
	return res, "";
}

func (s *SmartContract) getSumToRepaymentInner(stub shim.ChaincodeStubInterface, sum float64, percent float64) (float64) {
	return sum * percent;
}

func (s *SmartContract) getOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if args[1] == "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not be null orderId\"}")
	}
	order, _, err := s.getOrderByIdInner(stub, args[1]);
	if err != "" {
		logger.Error(err)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err+"\"}")
	}
	off, _ := json.Marshal(order)
	logger.Info("return data orderJson = "+string(off))
	return shim.Success(off)
}

func (s *SmartContract) activateOrder(stub shim.ChaincodeStubInterface, user User, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if args[1] == "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not be null orderId\"}")
	}
	order, key, err := s.getOrderByIdInner(stub, args[1]);
	if err != "" {
		logger.Error(err)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err+"\"}")
	}
	if order.Status == "ACTIVE" {
		logger.Info("########### "+projectName+" "+version+" success activateOrder ("+args[0]+") Already ACTIVE ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already activated\"}"))
	}
	if order.Status != "CONFIRMED" {
		logger.Error("Order could by in status 'CONFIRMED', current status '"+order.Status+"'")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Order could by in status 'CONFIRMED', current status '"+order.Status+"'\"}")
	}

	if(args[0] != order.Lender){
		if(user.Role == "ADMIN"){
			logger.Info("########### "+projectName+" "+version+" success activateOrder ("+args[0]+") by ADMIN ###########")
			//return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"(ADMIN) Order successfully activated\"}"));
		} else {
			logger.Error("Failed to activate ORDER ("+args[0]+") you can not activate order")
			jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to activate ORDER, need user '"+order.Lender+"'\"}"
			return shim.Error(jsonResp)
		}
	}
	order.Status = "ACTIVE"
	compositeKeyId, err2 := stub.CreateCompositeKey("ORDER-ACTIVE", []string{args[1]})
	logger.Info("compositeKey = ", compositeKeyId)
	if err2 != nil {
		logger.Error("Can't create siquence ORDER-ACTIVE for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create siquence ORDER-ACTIVE\"}")
	}
	err2 = stub.PutState(compositeKeyId, []byte(args[1]))
	if err2 != nil {
		logger.Error("Failed to activate siquence ORDER-ACTIVE ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to activate siquence ORDER-ACTIVE "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	ofr, _ := json.Marshal(order)
	err2 = stub.PutState(key, ofr)

	if err2 != nil {
		logger.Error("Failed to activate ORDER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to activate ORDER ("+args[0]+") "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success activateOrder ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order successfully activated\"}"));
}

func (s *SmartContract) getBorrowOffersIdsList(stub shim.ChaincodeStubInterface, key string, args []string) ([]string,string) {
	offers, err2 := stub.GetState(key)
	arr := []string{};
	if err2 != nil {
		logger.Errorf("Can't get ACTIVE_OFFERS array("+key+")")
		return arr,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get ACTIVE_OFFERS array("+key+")}";
	}
	err2 = json.Unmarshal([]byte(offers), &arr)
	if err2 != nil {
		logger.Errorf("Can't parse ACTIVE_OFFERS array("+string(offers)+")")
		return arr, ("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't parse ACTIVE_OFFERS array("+string(offers)+")}")
	}
	return arr, "";
}

func (s *SmartContract) getBorrowOffers(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, sumStart, percentStart, periodStart, sumFinish, percentFinish, periodFinish}
	return s.getOfferListInner(stub, "ACTIVE_OFFERS_BORROWS", args)
}


func (s *SmartContract) getLendOffers(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, sumStart, percentStart, periodStart, sumFinish, percentFinish, periodFinish}
	return s.getOfferListInner(stub, "ACTIVE_OFFERS_LENDS", args)
}

func (s *SmartContract) getOfferListInner(stub shim.ChaincodeStubInterface, key string, args []string) pb.Response {
	//{Login - 0, sumStart, percentStart, periodStart, sumFinish, percentFinish, periodFinish}

	if len(args) != 7 {
		logger.Error("Incorrect number of arguments. Expecting 7, current len = "+strconv.Itoa(len(args)))
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 4, current len = "+strconv.Itoa(len(args))+"\"}")
	}

	percentSt,err5 := strconv.ParseFloat(args[2], 32);
	if(err5 != nil){
		logger.Error("Can not parse Percent data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Percent data (need float32 value)\"}")
	}
	percentFin,err5 := strconv.ParseFloat(args[5], 32);
	if(err5 != nil){
		logger.Error("Can not parse Percent data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Percent data (need float32 value)\"}")
	}
	sumSt, err6 := strconv.ParseFloat(args[1], 64)
	if(err6 != nil){
		logger.Error("Can not parse Sum data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Sum data (need float value)\"}")
	}
	sumFin, err6 := strconv.ParseFloat(args[4], 64)
	if(err6 != nil){
		logger.Error("Can not parse Sum data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Sum data (need float value)\"}")
	}
	periodSt,err7 := strconv.Atoi(args[3]);
	if(err7 != nil){
		logger.Error("Can not parse Period data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Period data (need int value)\"}")
	}
	periodFin,err7 := strconv.Atoi(args[6]);
	if(err7 != nil){
		logger.Error("Can not parse Period data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Period data (need int value)\"}")
	}
	arr, err2 := s.getBorrowOffersIdsList(stub, key, args);
	if(err2 != ""){
		return shim.Error(err2)
	}

	var buffer bytes.Buffer
	buffer.WriteString("\"total\":"+strconv.Itoa(len(arr))+",\"result\":[")
	alreadyWritten := false
	count := 0;

	for i := range arr {
		order, _, err2 := s.getOrderByIdInner(stub, arr[i]);
		if(err2!=""){
			return shim.Error(err2)
		} else {
			if(order.Percent>=percentSt && order.Sum >=sumSt && order.Period>=periodSt &&
					order.Percent<=percentFin && order.Sum <=sumFin && order.Period<=periodFin){
				ordStr, _ := json.Marshal(order)
				if alreadyWritten == true {
					buffer.WriteString(",")
					alreadyWritten=false
				}
				buffer.WriteString(string(ordStr));
				alreadyWritten = true
				count++
			}
		}
	}
	buffer.WriteString("]}")
	result := "{\"length\":"+strconv.Itoa(count)+","+buffer.String();

	logger.Info("return data offersJson = "+result)
	logger.Info("########### "+projectName+" "+version+" success getOffers for user="+args[0]+" ###########")
	return shim.Success([]byte(result));
}

func (s *SmartContract) updateOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//{Login - 0, period, Borrower, Lender, Sum, Percent, StartDate, FinishDate, orderId, contractId}
	var types string
	if len(args) != 10 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 10, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if args[2] == "" && args[3] =="" {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null BORROWER='"+args[2]+"' and LENDER='"+args[3]+"'\"}")
	}
	if args[2] == args[3] {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be BORROWER=LENDER\"}")
	}

	percent,err5 := strconv.ParseFloat(args[5], 32);
	if(err5 != nil){
		logger.Error("Can not parse Percent data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Percent data (need float32 value)\"}")
	}
	period,err3 := strconv.Atoi(args[1]);
	if(err3 != nil){
		logger.Error("Can not parse period data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Period data (need int value)\"}")
	}
	sum, err6 := strconv.ParseFloat(args[4], 64)
	if(err6 != nil){
		logger.Error("Can not parse Sum data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Sum data (need float value)\"}")
	}
	contractId, err6 := strconv.ParseUint(args[9],10,64);
	if(err6 != nil){
		logger.Error("Can not parse contractId data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse contractId data (need uint64 value)\"}")
	}

	order, compositeKeyId, err2 := s.getOrderByIdInner(stub, args[8])
	logger.Info("compositeKey = ", compositeKeyId)
	if err2 != "" {
		logger.Error(err2)
		return shim.Error(err2)
	}

	if args[2]!= "" && len(order.Borrower)==0 {
		if args[2] == order.Lender {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be BORROWER=LENDER\"}")
		}
		order.Borrower = args[2];
		err4 := s.addOrderToUser(stub, "BORROWER", args[2], compositeKeyId)
		if err4 != "" {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err4+"\"}")
		}
	}
	if args[3]!="" && len(order.Lender)==0 {
		if args[3] == order.Borrower {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be BORROWER=LENDER\"}")
		}
		order.Lender = args[3];
		err4 :=s.addOrderToUser(stub, "LENDER", args[3], compositeKeyId)
		if err4 != "" {
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err4+"\"}")
		}
	}
	if contractId!=0 {
		order.ContractId=contractId
	}
	if period > 0 {
		order.Period = period;
	}
	if sum > 0 {
		order.Sum = sum;
	}
	if percent > 0 {
		order.Percent = percent;
	}
	if len(args[6])>0 {
		order.StartDate = args[6];
	}
	if len(args[7])>0 {
		order.FinishDate = args[7];
	}

	ord, _ := json.Marshal(order)
	logger.Info("return data orderJson = "+string(ord))
	err3 = stub.PutState(compositeKeyId, ord)
	if err3 != nil {
		logger.Errorf("Can't put data ORDER for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't put data ORDER\"}")
	}
	logger.Info("########### "+projectName+" "+version+" success updateOrder for user="+args[0]+" ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order for "+types+" added successfully\"}"))
}

func (s *SmartContract) closeOrder(stub shim.ChaincodeStubInterface, user User, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if args[1] == "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not be null orderId\"}")
	}
	order, key, err := s.getOrderByIdInner(stub, args[1]);
	if err != "" {
		logger.Error(err)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err+"\"}")
	}
	if order.Status == "CLOSED" {
		logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") Already CLOSED ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already closed\"}"))
	}
	if order.Status == "CANCELED" {
		logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") Already CANCELED ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already canceled\"}"))
	}
	if order.Status != "ACTIVE" {
		logger.Error("Order must be 'ACTIVE'")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order must be 'ACTIVE'\"}")
	}

	if(args[0] != order.Lender){
		if(user.Role == "ADMIN"){
			logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") by ADMIN ###########")
			//return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"(ADMIN) Order successfully closed\"}"));
		} else {
			logger.Error("Failed to CLOSE ORDER ("+args[0]+") you can not close order")
			jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to close ORDER, need user '"+order.Lender+"'\"}"
			return shim.Error(jsonResp)
		}
	}
	order.Status = "CLOSED"
	compositeKeyId, err2 := stub.CreateCompositeKey("ORDER-ACTIVE", []string{args[1]})
	logger.Info("compositeKey = ", compositeKeyId)
	if err2 != nil {
		logger.Error("Can't create siquence ORDER-ACTIVE for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create siquence ORDER-ACTIVE\"}")
	}
	err2 = stub.PutState(compositeKeyId, []byte(""))
	if err2 != nil {
		logger.Error("Failed to remove siquence ORDER-ACTIVE ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to remove siquence ORDER-ACTIVE "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}

	ofr, _ := json.Marshal(order)
	err2 = stub.PutState(key, ofr)

	if err2 != nil {
		logger.Error("Failed to close ORDER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to close ORDER ("+args[0]+") "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order successfully closed\"}"));
}

func (s *SmartContract) cancelOrder(stub shim.ChaincodeStubInterface, user User, args []string) pb.Response {
	//{Login - 0, orderId}
	if len(args) != 2 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if args[1] == "" {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can not be null orderId\"}")
	}
	order, key, err := s.getOrderByIdInner(stub, args[1]);
	if err != "" {
		logger.Error(err)
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\""+err+"\"}")
	}
	if order.Status == "CLOSED" {
		logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") Already CLOSED ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order already closed\"}"))
	}
	if order.Status == "ACTIVE" {
		logger.Error("Order can't be 'ACTIVE'")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order can't be 'ACTIVE'\"}")
	}

	if(args[0] != order.Borrower){
		if(user.Role == "ADMIN"){
			logger.Info("########### "+projectName+" "+version+" success closeOrder ("+args[0]+") by ADMIN ###########")
			//return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"(ADMIN) Order successfully closed\"}"));
		} else {
			usr := order.Borrower;
			if (len(usr) == 0) { usr = order.Lender}
			if (args[0] != usr) {
				logger.Error("Failed to CANCEL ORDER (" + args[0] + ") you can not cancel order")
				jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Failed to cancel ORDER, need user '" + usr + "'\"}"
				return shim.Error(jsonResp)
			}
		}
	}
	order.Status = "CANCELED"
	compositeKeyId, err2 := stub.CreateCompositeKey("ORDER-ACTIVE", []string{args[1]})
	logger.Info("compositeKey = ", compositeKeyId)
	if err2 != nil {
		logger.Error("Can't create siquence ORDER-ACTIVE for login="+args[0])
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create siquence ORDER-ACTIVE\"}")
	}
	err2 = stub.PutState(compositeKeyId, []byte(""))
	if err2 != nil {
		logger.Error("Failed to remove siquence ORDER-ACTIVE ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to remove siquence ORDER-ACTIVE "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}

	ofr, _ := json.Marshal(order)
	err2 = stub.PutState(key, ofr)

	if err2 != nil {
		logger.Error("Failed to cancel ORDER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to cancel ORDER ("+args[0]+") "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success cancelOrder ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Order successfully canceled\"}"));
}


