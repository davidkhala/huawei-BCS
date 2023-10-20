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

var querybalanceCmd = &cobra.Command{
	Use:   "querybalance",
	Short: "Query current balance ",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(userid) == 0 {
			cmd.Help()
			return nil
		}
		return query()
	},
}

func QueryBalanceCmd() *cobra.Command {
	querybalanceCmd.Flags().StringVarP(&userid, "userid", "u", "", "user id")
	querybalanceCmd.Flags().StringVarP(&propwd, "protectpwd", "p", "", "protect pwd")
	querybalanceCmd.Flags().StringVarP(&idchaincode, "idcc", "I", "IDChaincode", "idendity chaincode name")
	querybalanceCmd.Flags().StringVarP(&txchaincode, "txcc", "T", "TxChaincode", "transaction chaincode name")
	querybalanceCmd.Flags().StringVarP(&channelid, "channel", "C", "mychannel", "channel id")
	querybalanceCmd.Flags().StringVarP(&orgid, "orgid", "o", "org1", "organization id")
	querybalanceCmd.Flags().StringVarP(&conf, "config", "c", "./config_test.yaml", "configuration file path")
	querybalanceCmd.Flags().StringVarP(&encryptstrength, "encryptstrength", "s", "", "encrypt strength")

	return querybalanceCmd
}

func query() error {
	fmt.Println("userid: ", userid)
	fmt.Println("protectwd: ", propwd)
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
		fmt.Println("user is not exist ")
		return errors.New("user is not exist")
	}
	//read user
	f, err = os.OpenFile(filename, os.O_RDONLY, 0600) //打开文件
	defer f.Close()
	check(err)

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
		ChainCodeID:     txchaincode,
	}
	if err := setup.Initialize(); err != nil {
		fmt.Println("fail to init sdk: ", err.Error())
		return errors.New("fail to init sdk: " + err.Error())
	}

	//1.get balance
	setup.ChainCodeID = txchaincode
	transRec := sdk_client.TransRecord{}

	fmt.Println("query balance")

	resps, err := sdk_client.Query(setup, "QueryBalance", [][]byte{[]byte(addrA)})
	if err != nil {
		fmt.Println("Fail to query balance :", err.Error())
		return err
	}

	err = json.Unmarshal(resps[0].ProposalResponse.GetResponse().Payload, &transRec)
	if err != nil {
		fmt.Println("unmarshal query result error: ", err.Error())
		return err
	}

	curbalance, err := pswapi_sdk.Decrypt(transRec.Balance, userdata.PriKey, propwd)
	if err != nil {
		fmt.Println("sdk Decrypt error: ", err.Error())
		return err
	}

	fmt.Println("current balance:" + curbalance.String())

	return nil
}
