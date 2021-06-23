package main

import (
	"fmt"
	"os"

	"github.com/Gravity-Tech/solanoid/commands"
)

func main() {

	if err := commands.SolanoidCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
