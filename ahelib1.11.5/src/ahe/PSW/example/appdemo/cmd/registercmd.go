package cmd

import (
	pswapi_sdk "ahe/PSW/api/ahelib"
	"ahe/PSW/example/appdemo/sdk_client"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var userid string
var propwd string
var initbalance string
var idchaincode string
var txchaincode string
var channelid string
var orgid string
var conf string
var encryptstrength string

var RegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "register saas user",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(userid) == 0 || len(propwd) == 0 {
			cmd.Help()
			return nil
		}
		return register()
	},
}

func RegCmd() *cobra.Command {
	RegisterCmd.Flags().StringVarP(&userid, "userid", "u", "", "user id")
	RegisterCmd.Flags().StringVarP(&propwd, "protectpwd", "p", "", "protect pwd")
	RegisterCmd.Flags().StringVarP(&initbalance, "initbalance", "i", "0", "init initbalance")
	RegisterCmd.Flags().StringVarP(&idchaincode, "idcc", "I", "IDChaincode", "idendity chaincode name")
	RegisterCmd.Flags().StringVarP(&txchaincode, "txcc", "T", "TxChaincode", "transaction chaincode name")
	RegisterCmd.Flags().StringVarP(&channelid, "channel", "C", "mychannel", "channel id")
	RegisterCmd.Flags().StringVarP(&orgid, "orgid", "o", "org1", "organization id")
	RegisterCmd.Flags().StringVarP(&conf, "config", "c", "./config_test.yaml", "configuration file path")
	RegisterCmd.Flags().StringVarP(&encryptstrength, "encryptstrength", "s", "", "encrypt strength")

	return RegisterCmd
}

func check(e error) {
	if e != nil {
		fmt.Println(e.Error())
		panic(e)
	}
}

func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func register() error {
	fmt.Println("userid: ", userid)
	fmt.Println("protectwd: ", propwd)
	fmt.Println("initbalance: ", initbalance)
	fmt.Println("idcc: ", idchaincode)
	fmt.Println("txcc: ", txchaincode)
	fmt.Println("channelid: ", channelid)
	fmt.Println("orgid: ", orgid)
	fmt.Println("conf: ", conf)
	fmt.Println("encryptstrength: ", encryptstrength)

	var filename = userid + ".data"
	var f *os.File
	var err error
	//check file

	userdata := &UserData{}
	if encryptstrength != "false" {
		pswapi_sdk.SetEncryptStrength(encryptstrength)
	}

	//check user
	if Exists(filename) {
		fmt.Println("user already registered")
		return errors.New("user already registered")
	}

	privKeyStr, pubKeyStr, err := pswapi_sdk.GenerateKey(propwd)
	check(err)
	userdata.PubKey = pubKeyStr
	userdata.PriKey = privKeyStr

	//register
	senderAddr := calcAddr(userdata.PubKey)
	setup := &sdk_client.BaseSetupImpl{
		ConfigFile:      conf,
		ChannelID:       channelid,
		OrgID:           orgid,
		ConnectEventHub: false,
		ChainCodeID:     idchaincode,
	}
	if err := setup.Initialize(); err != nil {
		fmt.Println("fail to init sdk: ", err.Error())
		return errors.New("fail to init sdk: " + err.Error())
	}
	setup.ChainCodeID = idchaincode

	//	senderAddrByte, _ := hex.DecodeString(senderAddr)
	//	_, err = sdk_client.Invoke(setup, "Register", [][]byte{[]byte(senderAddr), []byte(userdata.PubKey)})
	_, err = sdk_client.Invoke(setup, "Register", [][]byte{[]byte(userdata.PubKey), []byte(senderAddr)})
	if err != nil {
		fmt.Println("Fail to register user pk ", err.Error())
		return errors.New("Fail to register user pk " + err.Error())
	}

	//addrByte := res[0].ProposalResponse.GetResponse().Payload
	//fmt.Println("Register pk: ", string(addrByte))
	fmt.Println("Register pk success: ", senderAddr)

	balanceInfo, err := pswapi_sdk.InitBalance(initbalance, userdata.PubKey)
	check(err)

	setup.ChainCodeID = txchaincode
	_, err = sdk_client.Invoke(setup, "init", [][]byte{[]byte(userdata.PubKey), []byte(balanceInfo)})
	if err != nil {
		fmt.Println("init balance error for user: ", senderAddr, err.Error())
		return errors.New("init balance error for user: " + senderAddr + err.Error())
	}

	fmt.Println("init balance successfully: ", senderAddr)

	// save user key
	f, err = os.Create(filename)
	defer f.Close()
	writebyte, err := json.Marshal(userdata)
	check(err)

	_, err1 := io.WriteString(f, string(writebyte))
	check(err1)

	return nil
}
