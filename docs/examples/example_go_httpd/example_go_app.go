package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Printf("before sleep\n")
	time.Sleep(time.Second)
	fmt.Printf("after sleep\n")
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "my first unikernel!")
}
