package cmd

import (
	"ahe/PSW/example/appdemo/sdk_client"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var SetEncryptStrengthCmd = &cobra.Command{
	Use:   "setstrength",
	Short: "set strength cmd",
	RunE: func(cmd *cobra.Command, args []string) error {
		return setEncryptStrength()
	},
}

func SetStrengthCmd() *cobra.Command {
	SetEncryptStrengthCmd.Flags().StringVarP(&encryptstrength, "encryptstrength", "s", "", "encryptstrength")
	SetEncryptStrengthCmd.Flags().StringVarP(&idchaincode, "idcc", "I", "IDChaincode", "idendity chaincode name")
	SetEncryptStrengthCmd.Flags().StringVarP(&txchaincode, "txcc", "T", "TxChaincode", "transaction chaincode name")
	SetEncryptStrengthCmd.Flags().StringVarP(&channelid, "channel", "C", "mychannel", "channel id")
	SetEncryptStrengthCmd.Flags().StringVarP(&orgid, "orgid", "o", "org1", "organization id")
	SetEncryptStrengthCmd.Flags().StringVarP(&conf, "config", "c", "./config_test.yaml", "configuration file path")

	return SetEncryptStrengthCmd
}

func setEncryptStrength() error {
	fmt.Println("encryptstrength: ", encryptstrength)
	fmt.Println("idcc: ", idchaincode)
	fmt.Println("txcc: ", txchaincode)
	fmt.Println("channel: ", channelid)
	fmt.Println("orgid: ", orgid)
	fmt.Println("conf: ", conf)

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

	_, err := sdk_client.Invoke(setup, "SetEncryptStrength", [][]byte{[]byte(encryptstrength)})
	if err != nil {
		fmt.Println("Fail to set length level :", err.Error())
		return err
	}
	return nil
}
