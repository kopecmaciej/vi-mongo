package main

import (
	"fmt"
	"os"

	"github.com/kopecmaciej/vi-mongo/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "\nERROR: Application crashed unexpectedly: %v\n", r)
			fmt.Fprintf(os.Stderr, "Please check the log file for details\n")
			os.Exit(1)
		}
	}()

	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
