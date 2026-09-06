package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/cwbudde/go-dws/cmd/dwscript/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if !errors.Is(err, cmd.ErrSilent) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
