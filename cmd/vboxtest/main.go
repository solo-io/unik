package main

import "os"

func main() {
	diskPath := "/Users/pivotal/VirtualBox VMs/Windows10/Windows10.vbox"
	baseFolder := os.Getenv("PWD")
	bridgeName := "bridgestuff"
	bridgeAdapterKey := 0
	diskFile := "/Users/pivotal/tmp/program.vmdk"
	//todo: get a program.vmdk
}