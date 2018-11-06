package main

import (
	"os"

	"github.com/Code-Hex/grpcrnd"
)

func main() {
	os.Exit(grpcrnd.New().Run())
}
