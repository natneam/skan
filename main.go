package main

import (
	"fmt"

	"natneam.com/skan/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Println(err)
	}
}
