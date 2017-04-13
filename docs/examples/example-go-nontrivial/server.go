package main

import (
	"fmt"

	"github.com/cf-unik/unik/docs/examples/example-go-nontrivial/func1"
	"github.com/cf-unik/unik/docs/examples/example-go-nontrivial/func2"
)

func main() {
	fmt.Printf(func1.Func1())
	fmt.Printf(func2.Func2())
}
