package main

import (
	pswapi_sdk "ahe/PSW/api/ahelib"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInit("1", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, addr string, plaintext int64, privkey string, protectedpwd string) {
	trnasRecbytes := stub.State[addr]
	if trnasRecbytes == nil {
		fmt.Println("State", addr, "failed to get value")
		t.FailNow()
	}
	transRecStruct := &TransRecord{}
	err := json.Unmarshal(trnasRecbytes, transRecStruct)
	if err != nil {
		t.Fatal("fail to unmarshal transRec")
	}

	plainText, err := pswapi_sdk.Decrypt(string(transRecStruct.Balance), privkey, protectedpwd)
	if err != nil {
		fmt.Println("decrypt error")
		t.FailNow()
	}

	result := new(big.Int).SetInt64(plaintext)

	if plainText.Cmp(result) != 0 {
		fmt.Println("State value: ", plainText.String())
		fmt.Println("result value", result.String())
		t.FailNow()

	}
}

func getHash(cont string) (string, error) {

	Hasher := crypto.SHA256.New()
	Hasher.Write([]byte(cont))
	HashRes := Hasher.Sum(nil)

	return hex.EncodeToString(HashRes), nil
}

func checkQuery(t *testing.T, stub *shim.MockStub, name string, value string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("query"), []byte(name)})
	if res.Status != shim.OK {
		fmt.Println("Query", name, "failed", string(res.Message))
		t.FailNow()
	}
	if res.Payload == nil {
		fmt.Println("Query", name, "failed to get value")
		t.FailNow()
	}
	if string(res.Payload) != value {
		fmt.Println("Query value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", "failed", string(res.Message))
		t.FailNow()
	} else {
		fmt.Println(string(args[0]), "res: ", string(res.Payload))
	}
}

func TestHeDemoChaincode_register(t *testing.T) {
	protectedpwd := "123456A"
	scc := new(TransChaincodeDemo)
	stub := shim.NewMockStub("TransChaincodeDemo", scc)

	checkInit(t, stub, [][]byte{[]byte("init"), []byte("A"), []byte("123"), []byte("B"), []byte("234")})

	privKeyA, pubKeyA, err := pswapi_sdk.GenerateKey(protectedpwd)
	if err != nil {
		t.Fatalf("fail to generate key for sender")
		t.FailNow()
	}

	hashAddrA, _ := getHash(pubKeyA)

	privKeyB, pubKeyB, err := pswapi_sdk.GenerateKey(protectedpwd)
	if err != nil {
		t.Fatalf("fail to generate key for receiver")
		t.FailNow()
	}

	hashAddrB, _ := getHash(pubKeyB)

	initBalanceInfoA, err := pswapi_sdk.InitBalance("100", pubKeyA)
	if err != nil {
		t.Fatal("fail to generate initbalance info")
	}

	initBalanceInfoB, err := pswapi_sdk.InitBalance("200", pubKeyB)
	if err != nil {
		t.Fatal("fail to generate initbalance info")
	}

	checkInvoke(t, stub, [][]byte{[]byte("init"), []byte(pubKeyA), []byte(initBalanceInfoA)})
	checkInvoke(t, stub, [][]byte{[]byte("init"), []byte(pubKeyB), []byte(initBalanceInfoB)})

	//Tx 1
	//get A's balance
	transRecABytes := stub.State[hashAddrA]
	transRecAStruct := &TransRecord{}
	err = json.Unmarshal(transRecABytes, transRecAStruct)
	if err != nil {
		t.Fatal("fail to unmarshal transRec")
	}
	cipherA := transRecAStruct.Balance
	//prepare a->b 10
	txInfo, err := pswapi_sdk.PrepareTxInfo(string(cipherA), "10", pubKeyA, pubKeyB, privKeyA, protectedpwd)
	if err != nil {
		t.Fatal("fail to prepare tx info: ", err.Error())
	}
	//send transaction
	checkInvoke(t, stub, [][]byte{[]byte("Transfer"), []byte(hashAddrA), []byte(hashAddrB), []byte(txInfo)})
	//check balance
	checkState(t, stub, hashAddrA, 90, privKeyA, protectedpwd)
	checkState(t, stub, hashAddrB, 210, privKeyB, protectedpwd)

	//Tx 2
	//get A's balance
	transRecABytes = stub.State[hashAddrA]
	err = json.Unmarshal(transRecABytes, transRecAStruct)
	if err != nil {
		t.Fatal("fail to unmarshal transRec")
	}
	cipherA = transRecAStruct.Balance
	//prepare a->b 10
	txInfo, err = pswapi_sdk.PrepareTxInfo(string(cipherA), "10", pubKeyA, pubKeyB, privKeyA, protectedpwd)
	if err != nil {
		t.Fatal("fail to prepare tx info")
	}
	//send transaction
	checkInvoke(t, stub, [][]byte{[]byte("Transfer"), []byte(hashAddrA), []byte(hashAddrB), []byte(txInfo)})
	//check balance
	checkState(t, stub, hashAddrA, 80, privKeyA, protectedpwd)
	checkState(t, stub, hashAddrB, 220, privKeyB, protectedpwd)

	//Tx 3
	//get A's balance
	transRecABytes = stub.State[hashAddrB]
	err = json.Unmarshal(transRecABytes, transRecAStruct)
	if err != nil {
		t.Fatal("fail to unmarshal transRec")
	}
	cipherA = transRecAStruct.Balance
	//prepare b->a 50
	txInfo, err = pswapi_sdk.PrepareTxInfo(string(cipherA), "50", pubKeyB, pubKeyA, privKeyB, protectedpwd)
	if err != nil {
		t.Fatal("fail to prepare tx info")
	}
	//send transaction
	checkInvoke(t, stub, [][]byte{[]byte("Transfer"), []byte(hashAddrB), []byte(hashAddrA), []byte(txInfo)})
	//check balance
	checkState(t, stub, hashAddrA, 130, privKeyA, protectedpwd)
	checkState(t, stub, hashAddrB, 170, privKeyB, protectedpwd)

}
