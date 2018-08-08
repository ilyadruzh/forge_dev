package main

import (
	"bytes"
	"encoding/json"
	"strconv"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

var projectName = "lend"
var version = "v1"
var logger = shim.NewLogger("loans")

type SmartContract struct {
}

type User struct {
	Login 			string `json:"login"`
	Pass 			string `json:"pass"`
	Patronymic 		string `json:"patronymic"`
	FirstName 		string `json:"firstName"`
	LastName 		string `json:"lastName"`
	BirthDate		string `json:"birthDate"`
	Phone			string `json:"mobilePhone"`
	Passport 		string `json:"passport"`
	Email 			string `json:"email"`
	Role 			string `json:"role"`
	Props			string `json:"props"`
	Wallet			string `json:"wallet"`
	Eth			string `json:"eth"`
	Balance		       float64 `json:"balance"`
	Gender			uint64 `json:"sexId"`
	Inn			string `json:"inn"`
	BureauConsentDate 	string `json:"bureauConsentDate"`
	Address 		string `json:"addressData"`
	RegistrationAddressData	string `json:"registrationAddressData"`
	ManagerId		uint64 `json:"managerId"`
	Status 			string `json:"status"`
	BrainyId		uint64 `json:"brainyId"`
	Orders	              []string `json:"orders"`
	Schedules	      []string `json:"schedules"`
}
type UserBalance struct {
	StartBalance    float64 `json:"startBalance"`
	Date		string `json:"date"`
	User		string `json:"user"`
	Increase 	float64 `json:"increase"`
}

type Order struct {
	Id	 	string `json:"id"`
	Borrower 	string `json:"borrowerId"`
	Lender 		string `json:"lenderId"`
	Sum	 	float64 `json:"sum"`
	Percent 	float64 `json:"percent"`
	Period		int    `json:"period"`   // in Month
	CreateDate	string `json:"createDate"`
	StartDate	string `json:"startDate"`
	FinishDate	string `json:"finishDate"`
	Status 		string `json:"status"`
	ContractId	uint64 `json:"contractId"`
	ScheduleIds	[]string `json:"scheduleIds"`
}

type Tranche struct {
	Id 		    uint64  `json:"id"`
	IssueDate	    string  `json:"issueDate"`
	RepaymentDate	    string  `json:"repaymentDate"`
	Principal	    float64 `json:"principal"`
	Interest	    float64 `json:"interest"`
	Lgot		    bool    `json:"lgot"`
	EachRepaymentFee    float64 `json:"eachRepaymentFee"`
	Rest		    float64 `json:"rest"`
	ScheduleId	    string  `json:"scheduleId"`
}

type Schedule struct {
	Id 		uint64   `json:"id"`
	CreationDate	int64    `json:"creationDate"`
	Amount		float64  `json:"amount"`
	ChargeIssueFee	bool     `json:"chargeIssueFee"`
	Issued		bool     `json:"issued"`
	ActiveBefore	string   `json:"activeBefore"`
	Tranches	[]string `json:"tranches"`
	OrderId 	string   `json:"orderId"`
}

type Transfer struct {
	Tx		string `json:"tx"`
	From		string `json:"from"`
	To		string `json:"to"`
	Amount		float64 `json:"amount"`
	OrderId 	string `json:"orderId"`
}

type Operation struct {
	Type_ 		string `json:"type"`	  // ADD, CONFIRM, CLOSE, STATUS
	Contract 	string `json:"contract"`
	InputData	string `json:"inputData"` // args array
	User		string `json:"user"`
	Date		string `json:"date"`
}

//------------------------------------------------------------------------------------
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("########### "+projectName+" "+version+" Init ###########")
	stub.PutState("TRANCHE-LAST-INDEX", []byte("0"))
	stub.PutState("ORDER-LAST-INDEX", []byte("0"))
	stub.PutState("ACTIVE_OFFERS_BORROWS", []byte("[]"))
	stub.PutState("ACTIVE_OFFERS_LENDS", []byte("[]"))
	stub.PutState("SCHEDULE-LAST-INDEX", []byte("0"))
	return shim.Success(nil)
}
//------------------------------------------------------------------------------------
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("----------------------------------------------------------")
	function, args := stub.GetFunctionAndParameters()
	user := User{}
	compositeKey := ""

	if len(args[0]) == 0 {
		logger.Error("login can't be null")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"login can't be null\"}"
		return shim.Error(jsonResp)
	}
	if function != "registerUser" {
		var usr []byte
		var err error
		compositeKey, err = stub.CreateCompositeKey("USER", []string{args[0]})
		logger.Info("compositeKey = ", compositeKey)
		if err != nil {
			logger.Error("Can't create composite key login="+args[0])
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create composite key for login="+args[0]+"\"}")
		}
		usr, err = stub.GetState(compositeKey)
		logger.Info("userJson = ", string(usr))
		if err != nil {
			logger.Error("Can't get user with login="+args[0])
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get user with login="+args[0]+"\"}")
		}
		if(len(usr)<3){
			logger.Error("Can't get user with login="+args[0]+", user not found!")
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"User not found!\"}")
		}
		err = json.Unmarshal(usr, &user)
		logger.Info("user = ", user)
		if err != nil {
			logger.Error("["+string(usr)+"] Failed to decode JSON: " + args[0])
			return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"["+string(usr)+"] Failed to decode JSON " + args[0] +"\"}")
		}

		if function == "getUser" {
			//userMap := s.structToMap(user)
			user.Pass = "*** SECURED ***"
			//delete(userMap, "Pass")
			//delete(userMap, "Orders")
			//delete(userMap, "Schedules")
			//user.Orders = []string{}
			//user.Schedules = []string{}

			usr, _ = json.Marshal(user)
			logger.Info("return data userJson = "+string(usr))
			logger.Info("########### "+projectName+" "+version+" success getUser ("+args[0]+") ###########")
			return shim.Success(usr);
		} else
		if function == "activateUser"{
			return s.activateUser(stub, user, compositeKey, args)
		} else
		if user.Status != "ACTIVE" {
			err2 := "You login is not activated"
			logger.Error(err)
			return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"" + err2 + "\"}")
		}
	}

	switch function {
	// User story
	case "registerUser":
		s.addOperation(stub, args, "ADD", "USER", function)
		return s.registerUser(stub, args)
	case "userLogin":
		return s.userLogin(stub, user, compositeKey, args)
	case "changeUserPassword":
		s.addOperation(stub, args, "CHANGE", "USER", function)
		return s.changeUserPassword(stub, user, compositeKey, args)
	case "changeUserBalance":
		s.addOperation(stub, args, "CHANGE_BALANCE", "USER", function)
		return s.changeUserBalance(stub, user, compositeKey, args)
	case "getHistoryUserBalance":
		return s.getHistoryUserBalance(stub, args)
	case "setUserEth":
		s.addOperation(stub, args, "SET_USER_ETH", "USER", function)
		return s.setUserEth(stub, user, compositeKey, args)
		// Payments
	case "tokenTransfer":
		s.addOperation(stub, args, "PAYMENT", "TRANSFER", function)
		return s.tokenTransfer(stub, args)
	case "getHistoryTransferToken":
		return s.getHistoryTransferToken(stub, args)
	case "closeOrder":
		s.addOperation(stub, args, "CLOSE", "ORDER", function)
		return s.closeOrder(stub, user, args)
	case "cancelOrder":
		s.addOperation(stub, args, "CANCEL", "ORDER", function)
		return s.cancelOrder(stub, user, args)
	case "updateOrder":
		s.addOperation(stub, args, "UPDATE", "ORDER", function)
		return s.updateOrder(stub, args)
	case "getBorrowOffers":
		return s.getBorrowOffers(stub, args)
	case "getLendOffers":
		return s.getLendOffers(stub, args)
		// Order story
	case "createOrder":
		s.addOperation(stub, args, "ADD", "ORDER", function)
		return s.createOrder(stub, args)
	case "getOrder":
		return s.getOrder(stub, args)
	case "activateOffer":
		s.addOperation(stub, args, "ACTIVATE", "OFFER", function)
		return s.activateOffer(stub, args)
	case "confirmOrder":
		s.addOperation(stub, args, "CONFIRM", "ORDER", function)
		return s.confirmOrder(stub, args)
	case "activateOrder":
		s.addOperation(stub, args, "ACTIVATE", "ORDER", function)
		return s.activateOrder(stub, user, args)
	case "getBorrows":
		return s.getBorrows(stub, args)
	case "getLends":
		return s.getLends(stub, args)
	//case "repaymentLoan":
	//	s.addOperation(stub, args, "CLOSE", "ORDER", function)
	//	return s.repaymentLoan(stub, args)
	//case "getSumToRepayment":
	//	return s.getSumToRepayment(stub, args)
		// Schedule story
	case "createSchedule":
		s.addOperation(stub, args, "CREATE", "SCHEDULE", function)
		return s.createSchedule(stub, args)
	case "setActiveSchedule":
		s.addOperation(stub, args, "SET_ACTIVE", "SCHEDULE", function)
		return s.setActiveSchedule(stub, args)
	case "getActiveSchedule":
		return s.getActiveSchedule(stub, args)
	case "getScheduleList":
		return s.getScheduleList(stub, args)
	case "getScheduleById":
		return s.getScheduleById(stub, args)
	case "getListOfTranchesByDate":
		return s.getListOfTranchesByDate(stub, args)
	case "paymentTranche":
		s.addOperation(stub, args, "PAYMENT_TRANCHE", "SCHEDULE", function)
		return s.paymentTranche(stub, args)
	case "getUserTranches":
		return s.getUserTranches(stub, user, args)

	// Util story
	case "putOther":
		s.addOperation(stub, args, "OTHER", "OPERATION", function)
		return s.putOther(stub, args)
	case "getOther":
		return s.getOther(stub, args)
	case "getOperationData":
		return s.getOperationData(stub, args)


	default:
		logger.Error("Unknown action, check the first argument, must be name of operation. But got: %v", args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Invalid Smart Contract operation name.\"}"
		return shim.Error(jsonResp)
	}
}
//------------------------------------------------------------------------------------
func (s *SmartContract) addOperation(stub shim.ChaincodeStubInterface, args []string, type_ string, contract string, function string) pb.Response {
	// 	last point - user
	//	addOperation(stub, args, "ADD", "ORDER")
	var JSONasBytes []byte
	var buffer bytes.Buffer
	var compositeKey string

	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for i := 1; i<len(args); i++{
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(args[i])
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	times := time.Now().String()
	var operation = Operation{
		type_,
		contract,
		buffer.String(),
		args[0],
		times,
		}

	JSONasBytes, err := json.Marshal(operation)
	if err != nil {
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't MARSHAL (addOperation) for method "+function+"\"}"
		return shim.Error(jsonResp)
	}

	// time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String();
	compositeKey = "OPER_"+s.timeStampStr()

	err = stub.PutState(compositeKey, JSONasBytes)
	if err != nil {
		logger.Errorf("Failed to addOperation ")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to addOperation\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success addOperation ("+function+") ###########")
	return shim.Success(nil);
}
//------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------
func (s *SmartContract) getOperationData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0 - login
	// 1 - filter (contract type)
	// 2 - periodStart
	// 3 - periodEnd
	if len(args) != 4 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 4, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	logger.Info("########### "+projectName+" "+version+" getOperationData ###########")
	var periodStart string;
	var periodFinish string;
	if(len(args[2])== 0){
		periodStart = strconv.Itoa(time.Date(2017, 10, 1, 0, 0, 0, 0, time.Local).Nanosecond())
	} else {
		periodStart = args[2]
	};
	if(len(args[3])== 0){
		periodFinish = s.timeStampStr()
	} else {
		periodFinish = args[3]
	};

	startKey := "OPER_"+periodStart
	finishKey :="OPER_"+periodFinish

	resultsIterator, err := stub.GetStateByRange(startKey, finishKey)
	if err != nil {
		logger.Errorf("Can't get by range data")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get by range data\"}"
		return shim.Error(jsonResp)
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("{\"start\":\""+periodStart+"\",\"finish\":\""+periodFinish+"\",\"result\":[")
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
			if alreadyWritten == true {
				buffer.WriteString(",")
			}
			alreadyWritten = false

			if (len(args[1]) > 0) {
				operation := Operation{}
				err := json.Unmarshal(queryResponseValue, &operation)
				logger.Info("operation = ", operation)

				if (err == nil && operation.Contract == args[1]) {
					buffer.WriteString(queryResponse)
					alreadyWritten = true
				}
			} else {
				buffer.WriteString(queryResponse)
				alreadyWritten = true
			}
		}
	}
	buffer.WriteString("]}")

	logger.Info("########### "+projectName+" "+version+" success getOperationData ("+args[0]+") ###########")
	return shim.Success(buffer.Bytes());
}


//------------------------------------------------------------------------------------
func (s *SmartContract) getHistoryTxJson(stub shim.ChaincodeStubInterface, key string) pb.Response {

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error("{\"status\":false,\"description\":\""+err.Error()+"\"}")
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("{\"status\":false,\"description\":\""+err.Error()+"\"}")
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"isDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	logger.Info("History: \n"+buffer.String())
	return shim.Success(nil);
}

//------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------
//------------------------------------------------------------------------------------

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		logger.Errorf("Error starting INSURANCE "+version+" chaincode: %s", err)
	}
}

