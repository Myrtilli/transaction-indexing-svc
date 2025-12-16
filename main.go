package main

import (
	"os"

	"github.com/Myrtilli/transaction-indexing-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
