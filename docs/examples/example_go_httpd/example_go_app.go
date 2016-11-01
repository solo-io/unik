package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func main() {
	fmt.Printf("%v\n", foo())
}

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func foo() error {
	fmt.Printf("trying to curl www.google.com\n")

	client := http.Client{
		Transport: &http.Transport{
			Dial: dialTimeout,
		},
	}
	req, err := http.NewRequest("GET", "http://www.google.com", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("response was: %v\n", string(data))

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "my first unikernel!")
}
