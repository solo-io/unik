package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// to exit the program with error use fatal method only!!

const BROADCAST_LISTENING_PORT = 9876

func fatal(m interface{}) {
	log.Fatal(m)
}

func main() {
	//make logs available via http request
	logs := &bytes.Buffer{}
	if err := teeStdout(logs); err != nil {
		fatal(err)
	}

	mux := http.NewServeMux()
	//serve logs
	mux.HandleFunc("/logs", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "logs: %s", string(logs.Bytes()))
	})
	log.Printf("starting log server\n")
	http.ListenAndServe(fmt.Sprintf(":%v", BROADCAST_LISTENING_PORT), mux)
}

func teeStdout(writer io.Writer) error {
	multi := io.MultiWriter(os.Stdout, writer)
	go func() {
		for {
			_, err := io.Copy(multi, os.Stdin)
			if err != nil {
				fatal("copying pipe reader to multi writer: " + err.Error())
			}
		}
	}()
	return nil
}
