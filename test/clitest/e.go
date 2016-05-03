package main

import (
	"log"
	"net/http"
	"fmt"
	"bufio"
	"net"
	"github.com/emc-advanced-dev/pkg/errors"
)

func main() {
	httpd()
}

func httpd() {
	if err := listIfaces(); err != nil {
		log.Printf("err printing ifaces: %v", err)
	}

	go func() {
		log.Println("testing internet connection")
		conn, err := net.Dial("tcp", "74.125.29.104:80")
		if err != nil {
			log.Printf("ERROR: failed to DIAL www.google.com: " + err.Error())
		} else {
			fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
			status, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				log.Printf("ERROR: read string fromom connection to www.google.com: " + err.Error())
			} else {
				log.Printf("response: %s", status)
			}
		}
	}()
	log.Println("Starting to listen!!!")
	log.Fatal(http.ListenAndServe(":8080", http.FileServer(http.Dir("/"))))
}

func listIfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return errors.New("retrieving network interfaces" + err.Error())
	}
	for _, iface := range ifaces {
		log.Printf("found an interface: %v\n", iface)
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			log.Printf("inspecting address: %v", addr)
			switch v := addr.(type) {
			case *net.IPNet:
				log.Printf("inspecting ip: %v", v)
			}
		}
	}
	return nil
}