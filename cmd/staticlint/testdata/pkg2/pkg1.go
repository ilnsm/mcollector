package main

import (
	"os"
)

func foo() {
	// No error here
	os.Exit(1)
}
