package main

import (
	"os"

	"github.com/eljosho/dms-manager/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
