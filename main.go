package main

import "github.com/mgerb/mgphoto/common"

var version = "undefined"

func init() {
	if version != "undefined" {
		println("mgphoto ", version, "\n")
	}
}

func main() {
	common.Start()
}
