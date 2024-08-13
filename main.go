package main

import (
	"os"

	"github.com/kopecmaciej/mongui/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
