package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	fmt.Printf("listening on port :8080")
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fmt.Printf("cwd: %s", wd)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "my first unikernel!")
}
