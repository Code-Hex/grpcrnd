package main

import (
	"os"
)

func main() {
	os.Exit(grpcrnd.New().Run())
}
