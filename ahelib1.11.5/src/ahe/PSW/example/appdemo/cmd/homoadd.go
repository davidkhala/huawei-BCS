package cmd

import (
	pswapi_sdk "ahe/PSW/api/ahelib"
	"ahe/PSW/example/appdemo/sdk_client"
	"errors"
	"fmt"
	"math/big"

	"github.com/spf13/cobra"
)

var num1 string
var num2 string

var homoaddCmd = &cobra.Command{
	Use:   "homoadd",
	Short: "homomorphic addition test",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(num1) == 0 || len(num2) == 0 {
			cmd.Help()
			return nil
		}
		return add()
	},
}

func HomoAddCmd() *cobra.Command {
	homoaddCmd.Flags().StringVarP(&txchaincode, "txcc", "T", "TxChaincode", "transaction chaincode name")
	homoaddCmd.Flags().StringVarP(&encryptstrength, "encryptstrength", "s", "", "encrypt strength")
	homoaddCmd.Flags().StringVarP(&channelid, "channel", "C", "mychannel", "channel id")
	homoaddCmd.Flags().StringVarP(&orgid, "orgid", "o", "org1", "organization id")
	homoaddCmd.Flags().StringVarP(&conf, "config", "c", "./config_test.yaml", "configuration file path")
	homoaddCmd.Flags().StringVarP(&num1, "num1", "a", "34", "number one")
	homoaddCmd.Flags().StringVarP(&num2, "num2", "b", "56", "number two")

	return homoaddCmd
}

func add() error {
	fmt.Println("num1: ", num1)
	fmt.Println("num2: ", num2)
	fmt.Println("txcc: ", txchaincode)
	fmt.Println("channel: ", channelid)
	fmt.Println("orgid: ", orgid)
	fmt.Println("conf: ", conf)
	fmt.Println("encryptstrength: ", encryptstrength)

	tmpass := "12345678A"
	if encryptstrength != "false" {
		pswapi_sdk.SetEncryptStrength(encryptstrength)
	}

	privKeyStr, pubKeyStr, err := pswapi_sdk.GenerateKey(tmpass)

	cipher1, err := pswapi_sdk.Encrypt(num1, pubKeyStr)
	if err != nil {
		fmt.Println("fail to encrypt num1: ", num1)
		return errors.New("fail to encrypt num1")
	}

	cipher2, err := pswapi_sdk.Encrypt(num2, pubKeyStr)
	if err != nil {
		fmt.Println("fail to encrypt num2: ", num2)
		return errors.New("fail to encrypt num2")
	}

	fmt.Println("encrypted num1: ", string(cipher1[:64]), "...")
	fmt.Println("encrypted num2: ", string(cipher2[:64]), "...")
	//transaction
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

	fmt.Println("invoke homoadd")
	resps, err := sdk_client.Invoke(setup, "HomoAdd", [][]byte{[]byte(cipher1), []byte(cipher2)})
	if err != nil {
		fmt.Println("Fail to invoke HomoAdd :", err.Error())
		return errors.New("Fail to invoke HomoAdd")
	}

	homoaddRes := resps[0].ProposalResponse.GetResponse().Payload
	fmt.Println("encrypted homoaddRes: ", string(homoaddRes[:64]), "...")

	plainRes, err := pswapi_sdk.Decrypt(string(homoaddRes), privKeyStr, tmpass)
	if err != nil {
		fmt.Println("sdk Decrypt error: ", err.Error())
		return errors.New("sdk Decrypt error")
	}

	fmt.Println("decrypted homoadd res: ", plainRes.String())

	Num1BI, res := new(big.Int).SetString(num1, 10)
	if res == false {
		fmt.Println("fail to set big number for num1: ", num1)
		return errors.New("fail to set big number")
	}

	Num2BI, res := new(big.Int).SetString(num2, 10)
	if res == false {
		fmt.Println("fail to set big number for num2: ", num2)
		return errors.New("fail to set big number")
	}

	expectRes := new(big.Int).Add(Num1BI, Num2BI)
	if expectRes.Cmp(plainRes) != 0 {
		fmt.Println("failed: homo addition result is not equal to expect value")
		return errors.New("failed: homo addition result is not equal to expect value")
	}

	fmt.Println("success")
	return nil
}
