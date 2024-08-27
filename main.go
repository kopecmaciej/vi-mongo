package main

import (
	"os"

	"github.com/kopecmaciej/vi-mongo/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
