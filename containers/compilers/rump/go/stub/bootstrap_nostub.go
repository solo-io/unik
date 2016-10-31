// +build nostub

package main

import "log"

func bootstrap() error {
	log.Printf("nostub specified, skipping bootstrap")
	return nil
}
