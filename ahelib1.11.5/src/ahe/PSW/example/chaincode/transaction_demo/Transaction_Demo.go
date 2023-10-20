package main

import (
	pswapi_cc "ahe/PSW/api/ChainCode"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

//var logger = util.GetLog("TransChaincode Demo")

var logger = shim.NewLogger("TransChaincode Demo")

var NEWBALANCE_A = "newBalanceA"
var NEWBALANCE_B = "newBalanceB"
var NEWTRANSACTION_A = "newTxA"
var NEWTRANSACTION_B = "newTxB"
var TX_PAY = "P"
var TX_RECEIVE = "R"
var TX_INIT = "I"

type TransChaincodeDemo struct{}

type TransRecord struct {
	FromAddr string
	ToAddr   string
	TXType   string
	Balance  string
	TX       string
	remark   string
}

func (t *TransChaincodeDemo) Init(stub shim.ChaincodeStubInterface) pb.Response {

	return shim.Success(nil)
}

func (t *TransChaincodeDemo) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("enter Invoke")
	function, args := stub.GetFunctionAndParameters()

	if function == "QueryBalance" {
		return t.queryBalance(stub, args)
	} else if function == "Transfer" {
		return t.transfer(stub, args)
	} else if function == "init" {
		return t.init(stub, args)
	} else if function == "HomoAdd" {
		return t.homoAdd(stub, args)
	} else if function == "SetEncryptStrength" {
		return t.setEncryptStrength(stub, args)
	}

	return shim.Error("Invalid invoke function name: " + function)
}

func (t *TransChaincodeDemo) setEncryptStrength(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		logger.Error("Incorrect number of arguments. expect 1  arguments")
		return shim.Error("Incorrect number of arguments. expect 1 arguments")
	}
	logger.Debug("enter SetEncryptStrength")

	encryptStrength := args[0]
	pswapi_cc.SetEncryptStrength(encryptStrength)
	return shim.Success([]byte("Success"))
}

func (t *TransChaincodeDemo) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("enter Transfer")

	if len(args) != 3 {
		logger.Error("Incorrect number of arguments. expect 3 arguments")
		return shim.Error("Incorrect number of arguments. expect 3 arguments")
	}

	AddrA := args[0]
	AddrB := args[1]
	txInfo := args[2]

	if strings.Compare(AddrA, AddrB) == 0 {
		logger.Error("A' addr is the same B'Addr")
		return shim.Error("A' addr is the same B'Addr")
	}

	// read a's trans record
	logger.Debugf("read sender: %s trans record", string(AddrA))
	transRecA, err := stub.GetState(AddrA)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if transRecA == nil {
		return shim.Error("Entity not found")
	}

	var transRecAStruct = TransRecord{}
	err = json.Unmarshal(transRecA, &transRecAStruct)
	if err != nil {
		logger.Error("fail to unmarshal user's trans record")
		return shim.Error("fail to unmarshal user's trans record")
	}

	// read b's trans record
	logger.Debugf("read receiver: %s trans record", string(AddrB))
	transRecB, err := stub.GetState(AddrB)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if transRecA == nil {
		return shim.Error("Entity not found")
	}

	var transRecBStruct = TransRecord{}
	err = json.Unmarshal(transRecB, &transRecBStruct)
	if err != nil {
		logger.Error("fail to unmarshal user's trans record")
		return shim.Error("fail to unmarshal user's trans record")
	}

	//validate transfer information
	logger.Debug("validate transfer information")
	logger.Debug("tx information: ")
	logger.Debugf("%+v\n", string(txInfo))
	cipherBalanceAKeyABlock := transRecAStruct.Balance
	cipherBalanceBKeyBBlock := transRecBStruct.Balance

	newCipherBalanceA, newCipherBalanceB, newCipherTxA, newCipherTxB, err := pswapi_cc.ValidateTxInfo(txInfo, cipherBalanceAKeyABlock, cipherBalanceBKeyBBlock)
	if err != nil {
		logger.Error("fail to validate trans information")
		return shim.Error("fail to validate trans information")
	}

	// update a's balance

	transRecAStruct.Balance = newCipherBalanceA
	transRecAStruct.TX = newCipherTxA
	transRecAStruct.TXType = "P"

	AvalbytesUpdate, err := json.Marshal(transRecAStruct)
	if err != nil {
		logger.Error("fail to marshal balance update info")
		return shim.Error("Marshal Error")
	}

	logger.Debug("update sender -> " + string(AvalbytesUpdate))
	err = stub.PutState(AddrA, AvalbytesUpdate)
	if err != nil {
		logger.Error("fail to store state: ", err.Error())
		return shim.Error(err.Error())
	}

	// update b's balance
	transRecBStruct.Balance = newCipherBalanceB
	transRecBStruct.TX = newCipherTxB
	transRecBStruct.TXType = "R"
	BvalbytesUpdate, err := json.Marshal(transRecBStruct)
	if err != nil {
		logger.Error("fail to marshal balance update info")
		return shim.Error("Marshal Error")
	}

	logger.Debug("update receiver -> " + string(BvalbytesUpdate))
	err = stub.PutState(AddrB, BvalbytesUpdate)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("Success"))
}

func (t *TransChaincodeDemo) init(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Debug("enter init Balance")
	if len(args) != 2 {
		logger.Error("parameters number is not correct")
		return shim.Error("parameters number is not correct")
	}

	PubKey := string(args[0])
	BalanceInfo := string(args[1])
	logger.Info("PubKey:", PubKey)
	logger.Info("BalanceInfo:", BalanceInfo)
	//PubKey := string(args[2])

	hashPubkey, err := t.calcAddr(PubKey)

	logger.Debug("encrypt initial balance")

	UserPubKey, err := stub.GetState(hashPubkey)
	if err != nil {
		logger.Error("Error on query addr")
		return shim.Error("fail to query addr")
	}

	if UserPubKey != nil {
		logger.Error("addr already register")
		return shim.Error("addr already register")
	}

	//validate balance
	cipherBalance, err := pswapi_cc.ValidateInitBalance(BalanceInfo, PubKey)
	if err != nil {
		logger.Error("fail to Validate InitBalance ")
		return shim.Error("fail toValidate InitBalance")
	}

	logger.Debug("prepare init balance record")
	TxRec := &TransRecord{}
	TxRec.FromAddr = "sys"
	TxRec.ToAddr = hashPubkey
	TxRec.Balance = string(cipherBalance)
	TxRec.TX = string(cipherBalance)
	TxRec.TXType = TX_INIT

	TxRecStore, err := json.Marshal(TxRec)
	if err != nil {
		logger.Error("fail to marshal trans record")
		return shim.Error("fail to marshal trans record")
	}

	logger.Debug("serialized tx rec: ", string(TxRecStore)[1:64], "...")
	logger.Debug("serialized tx rec length: ", len(TxRecStore))
	// store user's trans record
	err = stub.PutState(hashPubkey, []byte(TxRecStore))
	if err != nil {
		logger.Error("fail to store trans record")
		return shim.Error("fail to store trans record")
	}

	return shim.Success([]byte("Success"))
}

/*
query account's balance
*/
func (t *TransChaincodeDemo) queryBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		logger.Error("Incorrect number of arguments. Expecting addr to query")
		return shim.Error("Incorrect number of arguments. Expecting addr to query")
	}

	Addr := string(args[0])
	// Get the state from the ledger
	balance, err := stub.GetState(Addr)
	if err != nil {
		logger.Error("fail to get state for: ", Addr)
		return shim.Error("fail to get state for: " + Addr)
	}

	if balance == nil {
		logger.Error("nil amount for: ", Addr)
		return shim.Error("nil amount for: " + Addr)
	}

	logger.Debug("state for: ", Addr, "is: ", balance)
	return shim.Success(balance)
}

/*
call homomorphic addition function
*/
func (t *TransChaincodeDemo) homoAdd(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		logger.Error("Incorrect number of arguments. Expecting two cipher")
		return shim.Error("Incorrect number of arguments. Expecting two cipher")
	}

	cipher1 := string(args[0])
	cipher2 := string(args[1])
	// Get the state from the ledger
	homoaddRes, err := pswapi_cc.Add(cipher1, cipher2)
	if err != nil {
		logger.Error("error on homoadd", err.Error())
		return shim.Error("error on homoadd" + err.Error())
	}

	return shim.Success([]byte(homoaddRes))
}

/*
query account's history balance
*/

/*
func (t *TransChaincodeDemo) queryHistoryBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		logger.Error("Incorrect number of arguments. Expecting name of the person to query")
		return shim.Error("Incorrect number of arguments. Expecting name of the person to query")
	}

	key := args[0]

	resultsIterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		logger.Error("fail to get history state for: ", key, "error: ", err.Error())
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString("|")
		}

		item, _ := json.Marshal(queryResponse)
		buffer.Write(item)
		bArrayMemberAlreadyWritten = true
	}
	logger.Debug("query result: ", buffer.String())

	return shim.Success(buffer.Bytes())
}
*/

func (t *TransChaincodeDemo) calcAddr(cont string) (string, error) {

	Hasher := crypto.SHA256.New()
	Hasher.Write([]byte(cont))
	HashRes := Hasher.Sum(nil)

	return hex.EncodeToString(HashRes), nil
}

func main() {
	logger.Debug("start transaction chaincode")
	err := shim.Start(new(TransChaincodeDemo))
	if err != nil {
		logger.Error("Error starting Simple chaincode: ", err.Error())
	}
}
