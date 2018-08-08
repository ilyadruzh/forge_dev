package main

import (
	"encoding/json"
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"time"
)

// key USER{login}
// key USER_BAL{login}
// key USER_ETH{Eth}


//-----------------------------------------------------------------------------------
// -----------------------------  User story ----------------------------------------
//-----------------------------------------------------------------------------------
func (s *SmartContract) getUserByEthAddress(stub shim.ChaincodeStubInterface, ethAddr string, args []string) (User,[]byte,string) {
	logger.Info("########### "+projectName+" "+version+" getUserByEthAddress ###########")
	user := User{}
	compositeKey, err := stub.CreateCompositeKey("USER_ETH", []string{ethAddr})

	if err != nil {
		logger.Error("Can't create USER ETH siquence")
		jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can't create USER ETH siquence\"}"
		return user,[]byte(""),jsonResp;
	}
	usrKey, err := stub.GetState(compositeKey)
	if err != nil {
		logger.Error("Can't get key for user with key="+compositeKey)
		return user,[]byte(""),"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get key for user by address ETH\"}"
	}
	usr, err1 := stub.GetState(string(usrKey))
	logger.Info("userJson = ", string(usr))
	if err1 != nil {
		logger.Error("Can't get user with key="+compositeKey)
		return user,usrKey,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get user by address ETH\"}"
	}
	err = json.Unmarshal(usr, &user)
	logger.Info("user = ", user)
	if err != nil {
		logger.Error("["+string(usr)+"] Failed to decode JSON: " + args[0])
		return user,usrKey,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"["+string(usr)+"] can not find user in BC (USERETHkey=" + compositeKey+",usrKey="+string(usrKey) +")\"}"
	}

	logger.Info("########### "+projectName+" "+version+" success getUserByEthAddress ("+args[0]+") ###########")
	return user,usrKey, "";
}
//-----------------------------------------------------------------------------------
func (s *SmartContract) registerUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// 0  - login
	// 1  - pass
	// 2  - firstName
	// 3  - lastName
	// 4  - patronymic
	// 5  - birthDate
	// 6  - passport
	// 7  - email
	// 8  - mobilePhone
	// 9  - role
	// 10 - props
	// 11 - wallet
	// 12 - balance   - float64
	// 13 - gender    - uint32
	// 14 - inn
	// 15 - bureauConsentDate
	// 16 - addressData
	// 17 - registrationAddressData
	// 18 - managerId  - uint64
	// 19 - Eth address
	// 20 - BrainyId

	logger.Info("########### "+projectName+" "+version+" registerUser ###########")

	if len(args) != 21 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 21, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if len(args[9]) == 0 {
		logger.Error("role can't be null")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"role can't be null\"}"
		return shim.Error(jsonResp)
	} else if args[9] != "BORROWER" && args[9] != "LENDER" && args[9] != "BOTH" && args[9] != "ADMIN" {
		logger.Error("can't find role='"+args[9]+"'")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"can't find role='"+args[9]+"'\"}"
		return shim.Error(jsonResp)
	}
	gender, err6 := strconv.ParseUint(args[13],10,64);
	if(err6 != nil){
		logger.Error("Can not parse sexId data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse sexId data (need uint32 value)\"}")
	}
	managerId, err6 := strconv.ParseUint(args[18],10,64);
	if(err6 != nil){
		logger.Error("Can not parse managerId data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse managerId data (need uint64 value)\"}")
	}
	brainyId, err6 := strconv.ParseUint(args[20],10,64);
	if(err6 != nil){
		logger.Error("Can not parse brainyId data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse brainyId data (need uint64 value)\"}")
	}
	balance, err6 := strconv.ParseFloat(args[12], 64)
	if(err6 != nil){
		logger.Error("Can not parse Balance data")
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse Balance data (need float value)\"}")
	}

	compositeKey, err := stub.CreateCompositeKey("USER", []string{args[0]})
	logger.Info("compositeKey = "+compositeKey)
	if err != nil {
		logger.Error("Can't create user with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	usr, _ := stub.GetState(compositeKey)
	if usr != nil {
		logger.Error("Failed to add User ("+args[0]+") already exists!")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to add User ("+args[0]+") already exists!\"}"
		return shim.Error(jsonResp)
	}

	user := &User{Login: args[0],
		Pass: args[1],
		FirstName: args[2],
		LastName: args[3],
		Patronymic: args[4],
		BirthDate: args[5],
		Passport: args[6],
		Email: args[7],
		Phone: args[8],
		Role: args[9],
		Props: args[10],
		Wallet: args[11],
		Balance: balance,
		Gender: gender,
		Inn: args[14],
		BureauConsentDate: args[15],
		Address: args[16],
		RegistrationAddressData: args[17],
		Eth: args[19],
		ManagerId: managerId,
		BrainyId: brainyId,
		Status: "NEW",
		Orders: []string{},
		Schedules: []string{}}

	userAsBytes, err2 := json.Marshal(user)
	if err2 != nil {
		logger.Error("Marshal Error: "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Marshal Error ("+args[0]+") "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("user = ", user)
	logger.Info("userAsBytes = ", string(userAsBytes))
	err = stub.PutState(compositeKey, userAsBytes)

	if err != nil {
		logger.Error("Failed to add USER ("+args[0]+"): "+err.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to add USER ("+args[0]+") "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	compositeKeyBalance, err5 := stub.CreateCompositeKey("USER_BAL", []string{args[0]})
	if err5 != nil {
		logger.Error("Can't create user balance history key with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user balance history key with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	bal := &UserBalance{
		StartBalance: balance,
		Date: time.Now().String(),
		User: args[0],
	}
	balAsBytes, err2 := json.Marshal(bal)
	if err2 != nil {
		logger.Error("Marshal Error balance: "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Marshal Error balance ("+args[0]+") "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	err = stub.PutState(compositeKeyBalance, balAsBytes)
	if err != nil {
		logger.Error("Failed to add user balance history ("+args[0]+"): "+err.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to add user balance history ("+args[0]+") "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	if(args[11] != ""){
		res := s.addUserEth(stub, args[11], compositeKey, args);
		if(res != ""){
			return shim.Error(res);
		}
	}
	logger.Info("########### "+projectName+" "+version+" success addUser ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"User is successfully created\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) addUserEth(stub shim.ChaincodeStubInterface, ethAddr string, key string,args []string) string {

	//USER_ETH{Eth}
	ethKey, err := stub.CreateCompositeKey("USER_ETH", []string{ethAddr})

	if err != nil {
		logger.Error("Can't create USER ETH siquence")
		jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Can't create USER ETH siquence\"}"
		return jsonResp
	}
	err = stub.PutState(ethKey, []byte(key))
	if err != nil {
		logger.Error("Failed to add USER ETH siquence: " + err.Error())
		jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Failed to add USER ETH siquence\"}"
		return jsonResp
	}
	return ""
}

//------------------------------------------------------------------------------------
func (s *SmartContract) changeUserPassword(stub shim.ChaincodeStubInterface, user User, key string, args []string) pb.Response {

	logger.Info("########### "+projectName+" "+version+" changeUserPassword ("+args[0]+") ###########")
	// 0 - login
	// 1 - old hashPass
	// 2 - new hashPass
	if len(args) != 3 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = "+strconv.Itoa(len(args))+"\"}")

	}
	if user.Pass != args[1] {
		logger.Error("Password mismatch of: " + args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Password mismatch of " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}

	user.Pass = args[2]
	usr, _ := json.Marshal(user)
	err2 := stub.PutState(key, usr)

	if err2 != nil {
		logger.Error("Failed changing password USER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed changing password USER ("+args[0]+") "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success changeUserPassword ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Password is successfully changed\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getUserInner(stub shim.ChaincodeStubInterface, args []string) (User,string,string) {
	user := User{}
	var usr []byte

	compositeKey, err := stub.CreateCompositeKey("USER", []string{args[0]})
	logger.Info("compositeKey = ", compositeKey)
	if err != nil {
		logger.Error("Can't create siquence login="+args[0])
		return user,compositeKey,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create siquence for login="+args[0]+"\"}"
	}
	usr, err = stub.GetState(compositeKey)
	logger.Info("userJson = ", string(usr))
	if err != nil {
		logger.Error("Can't get user with login="+args[0])
		return user,compositeKey,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get user with login="+args[0]+"\"}"
	}
	err = json.Unmarshal(usr, &user)
	logger.Info("user = ", user)
	if err != nil {
		logger.Error("["+string(usr)+"] Failed to decode JSON: " + args[0])
		return user,compositeKey,"{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"["+string(usr)+"] Failed to decode JSON of " + args[0] +"\"}"
	}

	return user,compositeKey,"";
}
//------------------------------------------------------------------------------------
func (s *SmartContract) userLogin(stub shim.ChaincodeStubInterface, user User, key string, args []string) pb.Response {

	logger.Info("########### "+projectName+" "+version+" login ("+args[0]+") ###########")
	// 1 - login
	// 2 - hashPass
	if len(args) != 2 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = "+strconv.Itoa(len(args))+"\"}")
	}

	if user.Pass != args[1] {
		logger.Error("Password mismatch of: " + args[0]+"["+user.Pass+"]")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Password mismatch of " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	if user.Status != "ACTIVE" {
		logger.Error("You must activate your login: " + args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"You must activate your login " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success login ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"User is successfully login\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) activateUser(stub shim.ChaincodeStubInterface, user User, key string, args []string) pb.Response {

	logger.Info("########### "+projectName+" "+version+" activateUser ###########")
	// 1 - login
	// 2 - hashPass
	if len(args) != 2 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 2, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	if user.Pass != args[1] {
		logger.Error("Password mismatch of: " + args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Password mismatch of " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	if user.Status == "ACTIVE" {
		logger.Info("########### "+projectName+" "+version+" success activateUser ("+args[0]+") Already ACTIVE ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"User is already activated\"}"))
	}
	user.Status = "ACTIVE"
	usr, _ := json.Marshal(user)
	err2 := stub.PutState(key, usr)

	if err2 != nil {
		logger.Error("Failed to activate USER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed to activate USER ("+args[0]+") "+err2.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success activateUser ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"User is successfully activated\"}"));
}
//------------------------------------------------------------------------------------
func (s *SmartContract) changeUserBalance(stub shim.ChaincodeStubInterface, user User, key string, args []string) pb.Response {

	logger.Info("########### " + projectName + " " + version + " changeUserBalance (" + args[0] + ") ###########")
	// 0 - login
	// 1 - hashPass
	// 2 - change sum
	if len(args) != 3 {
		return shim.Error("{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = " + strconv.Itoa(len(args)) + "\"}")
	}
	if user.Pass != args[1] {
		logger.Error("Password mismatch of: " + args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Password mismatch of " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}

	res := s.changeUserBalanceInner(stub, user, key, args);
	if(res!=""){
		return shim.Error(res)
	} else {
		logger.Info("########### "+projectName+" "+version+" success changeUserBalance ("+args[0]+") ###########")
		return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Balance is successfully changed\"}"));
	}
}


func (s *SmartContract) changeUserBalanceInner(stub shim.ChaincodeStubInterface, user User, key string, args []string) string {

	increaseSum, err6 := strconv.ParseFloat(args[2], 64)
	if(err6 != nil){
		logger.Error("Can not parse increase sum data")
		return "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not parse increase sum data (need float value)\"}"
	}

	user.Balance = user.Balance + increaseSum;
	usr, _ := json.Marshal(user)
	err2 := stub.PutState(key, usr)

	if err2 != nil {
		logger.Error("Failed changing password USER ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed changing password USER ("+args[0]+") "+err2.Error()+"\"}"
		return jsonResp
	}

	// change balance history
	compositeKeyBalance, err5 := stub.CreateCompositeKey("USER_BAL", []string{args[0]})
	if err5 != nil {
		logger.Error("Can't create user balance history key with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user balance history key with login="+args[0]+"\"}"
		return jsonResp
	}
	bal, _ := stub.GetState(compositeKeyBalance)
	balanceHistory := []UserBalance{}

	if(bal != nil && len(string(bal))>0) {
		err3 := json.Unmarshal(bal, &balanceHistory)
		if err3 != nil {
			logger.Error("Error unmarshal balance: " + err3.Error())
			jsonResp := "{\"login\":\"" + args[0] + "\",\"status\":false,\"description\":\"Error unmarshal balance (" + args[0] + ") " + err3.Error() + "\"}"
			return jsonResp
		}
	}
	balObj := &UserBalance{
		Increase: increaseSum,
		Date: time.Now().String(),
		User: args[0],
	}
	balanceHistory = append(balanceHistory, *balObj)
	bal, _ = json.Marshal(balanceHistory)
	err4 := stub.PutState(compositeKeyBalance, bal)
	if err4 != nil {
		logger.Error("Failed change balance history ("+args[0]+"): "+err2.Error())
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Failed change balance history ("+args[0]+") "+err2.Error()+"\"}"
		return jsonResp
	}
	return "";
}
//------------------------------------------------------------------------------------
func (s *SmartContract) getHistoryUserBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("########### "+projectName+" "+version+" getHistoryUserBalance ("+args[0]+") ###########")
	// 0 - login
	if len(args) != 1 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 1, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	compositeKeyBalance, err5 := stub.CreateCompositeKey("USER_BAL", []string{args[0]})
	if err5 != nil {
		logger.Error("Can't create user balance history key with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user balance history key with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	bal, _ := stub.GetState(compositeKeyBalance)

	logger.Info("########### "+projectName+" "+version+" success getHistoryUserBalance ("+args[0]+") ###########")
	return shim.Success(bal);
}
//------------------------------------------------------------------------------------
func (s *SmartContract) setUserEth(stub shim.ChaincodeStubInterface, user User, key string, args []string) pb.Response {
	logger.Info("########### "+projectName+" "+version+" setUserEth ("+args[0]+") ###########")
	// 0 - login, 1 - wallet, 2 - addrEth
	if len(args) != 3 {
		return shim.Error("{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Incorrect number of arguments. Expecting 3, current len = "+strconv.Itoa(len(args))+"\"}")
	}
	compositeKey, err := stub.CreateCompositeKey("USER", []string{args[0]})
	logger.Info("compositeKey = "+compositeKey)
	if err != nil {
		logger.Error("Can't create user siquence with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user siquence with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	user.Wallet = args[1];
	user.Eth = args[2];
	usr, _ := json.Marshal(user)
	err2 := stub.PutState(key, usr)

	if err2 != nil {
		logger.Error("Filed change User Eth data")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Filed change User Eth data\"}"
		return shim.Error(jsonResp)
	}
	if(args[1] != ""){
		res := s.addUserEth(stub, args[1], compositeKey, args);
		if(res != ""){
			return shim.Error(res);
		}
	} else {
		logger.Error("Can not be null User Ethereum address")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can not be null User Ethereum address\"}"
		return shim.Error(jsonResp)
	};
	logger.Info("########### "+projectName+" "+version+" success setUserEth ("+args[0]+") ###########")
	return shim.Success([]byte("{\"login\":\""+args[0]+"\",\"status\":true,\"description\":\"Address ETH added successfully\"}"));
}