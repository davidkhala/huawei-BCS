package main

import (
	"ahe/PSW/example/appdemo/cmd"
	"fmt"
	"os"
)

func main() {
	cmd.Init()
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	var err error
	//setup, err = sdk_client.Init()
	if err != nil {

	}

}
