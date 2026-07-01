package main

import (
	"fmt"
	"os"

	"natneam.com/skan/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
