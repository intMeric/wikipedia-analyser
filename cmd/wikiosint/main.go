// cmd/wikiosint/main.go
package main

import (
	"fmt"
	"os"

	"github.com/intMeric/wikipedia-analyser/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
