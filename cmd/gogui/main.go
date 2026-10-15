package main

import (
	"fmt"
	"os"

	"github.com/menghun3-cn/go-gui-template/internal/app"
	"github.com/menghun3-cn/go-gui-template/internal/update"
)

var version = "dev"

func main() {
	root := app.New(version)
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "update" {
		if err := update.RunCLI(args[1:], version); err != nil {
			fmt.Fprintf(os.Stderr, "update: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := root.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
