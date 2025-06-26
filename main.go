package main

import (
	"github.com/go-johnnyhe/waveland/cmd"
)	

var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
}
