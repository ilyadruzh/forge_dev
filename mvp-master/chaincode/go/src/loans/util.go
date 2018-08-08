package main


import (
	"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"time"
	"net/url"
	"reflect"
)


func (s *SmartContract) putOther(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("########### " + projectName + " " + version + " putOther ###########")
	// login, key, value
	var user User
	var usr []byte

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3, current len = " + strconv.Itoa(len(args)))
	}
	compositeKey, err := stub.CreateCompositeKey("USER", []string{args[0]})
	logger.Info("compositeKey = ", compositeKey)
	if err != nil {
		logger.Errorf("Can't create user with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't create user with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	usr, err = stub.GetState(compositeKey)
	logger.Info("userJson = ", string(usr))
	if err != nil {
		logger.Errorf("Can't get user with login="+args[0])
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"Can't get user with login="+args[0]+"\"}"
		return shim.Error(jsonResp)
	}
	err = json.Unmarshal([]byte(usr), &user)
	if(user.Role != "ADMIN"){
		logger.Errorf("You can't put data, you need role=\"ADMIN\" to add, current role=\""+user.Role+"\"")
		jsonResp := "{\"login\":\""+args[0]+"\",\"status\":false,\"description\":\"You can't put data, you need role='ADMIN' to add, current role='"+user.Role+"'\"}"
		return shim.Error(jsonResp)
	}

	err = stub.PutState(args[1], []byte(args[2]));
	if err != nil {
		logger.Errorf("Failed to add other information ("+args[0]+"): "+err.Error())
		jsonResp := "{\"status\":false,\"description\":\"Failed to put others ("+args[0]+"): "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success putOther ("+args[0]+") ###########")
	return shim.Success([]byte("{\"status\":true,\"description\":\"Data added successfully\"}"));
}

func (s *SmartContract) getOther(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("########### " + projectName + " " + version + " getOther ###########")
	// login, key, value

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2, current len = " + strconv.Itoa(len(args)))
	}
	data, err := stub.GetState(args[1]);
	if err != nil {
		logger.Errorf("Failed to add other information ("+args[0]+"): "+err.Error())
		jsonResp := "{\"status\":false,\"description\":\"Failed to get others ("+args[0]+"): "+err.Error()+"\"}"
		return shim.Error(jsonResp)
	}
	logger.Info("########### "+projectName+" "+version+" success getOther ("+args[0]+") ###########")
	return shim.Success(data);
}

func (s *SmartContract) timeStampStr() string {
	return strconv.FormatInt(s.timeStamp(), 10)
}

func (s *SmartContract) timeStamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond);
}


func (s *SmartContract) appendData(arr []string,val string) []string {
	bool := false
	for i:= range arr {
		if arr[i]==val {
			bool = true;
		}
	}
	if bool == false {
		arr = append(arr, val)
	}
	return arr;
}

func (s *SmartContract) delElem(arr []string, value string) []string {
	c:= []string{}
	for i := range arr {
		if (arr[i] != value) {
			c = append(c, arr[i])
		}
	}
	return c;
}

func (s *SmartContract) structToMap(i interface{}) (values url.Values)  {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		// You ca use tags here...
		// tag := typ.Field(i).Tag.Get("tagname")
		// Convert each type into a string for the url.Values string map
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		}
		values.Set(typ.Field(i).Name, v)
	}
	return
}