package main

import (
	"github.com/go-johnnyhe/leetmock/cmd"
)	

var version = "dev"

func main() {
	cmd.Version = version
	cmd.Execute()
}
