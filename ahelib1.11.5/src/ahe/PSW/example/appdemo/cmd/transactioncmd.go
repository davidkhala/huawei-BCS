package cmd

import (
	pswapi_sdk "ahe/PSW/api/ahelib"
	"ahe/PSW/example/appdemo/sdk_client"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

var addrB string
var tx string

var TransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "transaction",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(addrB) == 0 {
			cmd.Help()
			return nil
		}
		return transaction()
	},
}

func TransCmd() *cobra.Command {
	TransactionCmd.Flags().StringVarP(&userid, "userid", "u", "", "user id")
	TransactionCmd.Flags().StringVarP(&encryptstrength, "encryptstrength", "s", "", "encrypt strength")
	TransactionCmd.Flags().StringVarP(&propwd, "protectpwd", "p", "", "protect pwd")
	TransactionCmd.Flags().StringVarP(&addrB, "AddrB", "b", "", "B' addr")
	TransactionCmd.Flags().StringVarP(&tx, "Tx", "t", "0", "Transaction Num")
	TransactionCmd.Flags().StringVarP(&idchaincode, "idcc", "I", "IDChaincode", "idendity chaincode name")
	TransactionCmd.Flags().StringVarP(&txchaincode, "txcc", "T", "TxChaincode", "transaction chaincode name")
	TransactionCmd.Flags().StringVarP(&channelid, "channel", "C", "mychannel", "channel id")
	TransactionCmd.Flags().StringVarP(&orgid, "orgid", "o", "org1", "organization id")
	TransactionCmd.Flags().StringVarP(&conf, "config", "c", "./config_test.yaml", "configuration file path")

	return TransactionCmd
}

func transaction() error {
	fmt.Println("userid: ", userid)
	fmt.Println("protectwd: ", propwd)
	fmt.Println("AddrB: ", addrB)
	fmt.Println("Tx: ", tx)
	fmt.Println("idcc: ", idchaincode)
	fmt.Println("txcc: ", txchaincode)
	fmt.Println("channel: ", channelid)
	fmt.Println("orgid: ", orgid)
	fmt.Println("conf: ", conf)
	fmt.Println("encryptstrength: ", encryptstrength)

	var filename = userid + ".data"
	var f *os.File
	var err error

	userdata := &UserData{}
	if encryptstrength != "false" {
		pswapi_sdk.SetEncryptStrength(encryptstrength)
	}

	if !Exists(filename) {
		fmt.Println("user is not exist")
		return errors.New("user is not exist")
	}
	//read user
	f, err = os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		fmt.Println("fail to open file: ", filename)
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	check(err)

	err = json.Unmarshal(b, userdata)
	check(err)

	//transaction
	addrA := calcAddr(userdata.PubKey)
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

	//1.get balance
	setup.ChainCodeID = txchaincode
	transRec := sdk_client.TransRecord{}

	resps, err := sdk_client.Query(setup, "QueryBalance", [][]byte{[]byte(addrA)})
	if err != nil {
		fmt.Println("Fail to query balance of sender: ", err.Error())
		return err
	}

	err = json.Unmarshal(resps[0].ProposalResponse.GetResponse().Payload, &transRec)
	if err != nil {
		fmt.Println("fail to unmarshal balance result: ", err.Error())
		return err
	}

	fmt.Println("get A's balance successfully")

	//2.get B's public key
	var pubKeyB string

	setup.ChainCodeID = idchaincode

	resps, err = sdk_client.Query(setup, "QueryPubkey", [][]byte{[]byte(addrB)})

	if err != nil {
		fmt.Println("Fail to query pubkey of receiver: ", err.Error())
		return errors.New("Fail to query pubkey of receiver: " + err.Error())
	}

	pubKeyB = string(resps[0].ProposalResponse.GetResponse().Payload)
	fmt.Println("Get B's ID successfully")

	//3.prepare
	cipherBalanceAKeyA := transRec.Balance
	txInfoSer, err := pswapi_sdk.PrepareTxInfo(cipherBalanceAKeyA, tx, userdata.PubKey, pubKeyB, userdata.PriKey, propwd)
	if err != nil {
		fmt.Println("fail to prepare tx info: ", err.Error())
		return errors.New("fail to prepare tx info: " + err.Error())
	}
	fmt.Println("prepare transaction information successfully")

	//4.invoke transaction

	setup.ChainCodeID = txchaincode
	_, err = sdk_client.Invoke(setup, "Transfer", [][]byte{[]byte(addrA), []byte(addrB), []byte(txInfoSer)})

	if err != nil {
		fmt.Println("Invoke Transfer error for user: ", addrA, err.Error())
		return errors.New("Invoke Transfer error for user: " + addrA + err.Error())
	}

	fmt.Println("Transfer success: ", addrA)

	return nil
}
