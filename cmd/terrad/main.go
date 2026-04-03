package main

import (
	"os"

	doapp "github.com/Daviddochain/dochain-core/v4/app"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func init() {}

func main() {
	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", doapp.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
