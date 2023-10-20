package cmd

import (
	"crypto"
	"encoding/hex"

	"github.com/op/go-logging"
	"github.com/spf13/cobra"
)

var logger = logging.MustGetLogger("appdemocmd")

var RootCmd = &cobra.Command{
	Use: "appdemo",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Init() {
	RootCmd.AddCommand(SetStrengthCmd())
	RootCmd.AddCommand(TransCmd())
	RootCmd.AddCommand(RegCmd())
	RootCmd.AddCommand(QueryBalanceCmd())
	RootCmd.AddCommand(HomoAddCmd())

}

type UserData struct {
	PriKey string
	PubKey string
}

func calcAddr(cont string) string {

	Hasher := crypto.SHA256.New()
	Hasher.Write([]byte(cont))
	HashRes := Hasher.Sum(nil)

	return hex.EncodeToString(HashRes)
}
