package main

import "C"
import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"unsafe"
)

//export gomaincaller
func gomaincaller(kludge_dns_addrs_len C.int, kludge_dns_addrs unsafe.Pointer, argc C.int, argv unsafe.Pointer) {
	os.Args = nil
	argcint := int(argc)
	argvarr := ((*[1 << 30]*C.char)(argv))
	for i := 0; i < argcint; i += 1 {
		os.Args = append(os.Args, C.GoString(argvarr[i]))
	}

	//cDnsAddrArray := ((*[1 << 30]*C.uint)(kludge_dns_addrs))
	dnsAddrs := C.GoBytes(kludge_dns_addrs, kludge_dns_addrs_len)

	if err := stub(dnsAddrs); err != nil {
		log.Printf("fatal: stub failed: %v", err)
		return
	}
	main()
}

const BROADCAST_LISTENING_PORT = 9967

func stub(dnsAddrs []byte) error {
	if len(dnsAddrs)%4 != 0 {
		errMsg := fmt.Sprintf("expected len(dnsAddrs) to be a multiple of 4, but instead got %v", dnsAddrs)
		return errors.New(errMsg)
	}
	var resolvConf string
	numAddrs := len(dnsAddrs) / 4
	for i := 0; i < numAddrs; i++ {
		b1 := dnsAddrs[i+0]
		b2 := dnsAddrs[i+1]
		b3 := dnsAddrs[i+2]
		b4 := dnsAddrs[i+3]
		resolvConf += fmt.Sprintf("nameserver %d.%d.%d.%d\n", b1, b2, b3, b4)
	}

	log.Printf("writing dns addr: %s from %v", resolvConf, dnsAddrs)

	if err := ioutil.WriteFile("/etc/resolv.conf", []byte(resolvConf), 0644); err != nil {
		return errors.New("filling in dns address " + err.Error())
	}

	//make logs available via http request
	logs := &bytes.Buffer{}
	if err := teeStdout(logs); err != nil {
		return errors.New("teeing stdout: " + err.Error())
	}
	if err := teeStderr(logs); err != nil {
		return errors.New("teeing stderr: " + err.Error())
	}
	log.SetOutput(os.Stderr)

	log.Printf("unik v0.0 boostrapping beginning...")

	if err := os.Chdir("/bootpart"); err != nil {
		return errors.New("changing wd to /bootpart: " + err.Error())
	}

	mux := http.NewServeMux()
	//serve logs
	mux.HandleFunc("/logs", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "logs: %s", string(logs.Bytes()))
	})
	log.Printf("starting log server\n")
	go func() {
		log.Printf("serving logs failed: %v", http.ListenAndServe(fmt.Sprintf(":%v", BROADCAST_LISTENING_PORT), mux))
	}()

	if err := bootstrap(); err != nil {
		return errors.New("bootstrap failed: " + err.Error())
	}
	return nil
}

func setEnv(env map[string]string) error {
	for key, val := range env {
		os.Setenv(key, val)
	}
	return nil
}

func teeStdout(writer io.Writer) error {
	r, w, err := os.Pipe()
	if err != nil {
		return errors.New("creating pipe: " + err.Error())
	}
	stdout := os.Stdout
	os.Stdout = w
	multi := io.MultiWriter(stdout, writer)
	reader := bufio.NewReader(r)
	go func() {
		for {
			_, err := io.Copy(multi, reader)
			if err != nil {
				panic("copying pipe reader to multi writer: " + err.Error())
			}
		}
	}()
	return nil
}

func teeStderr(writer io.Writer) error {
	r, w, err := os.Pipe()
	if err != nil {
		return errors.New("creating pipe: " + err.Error())
	}
	stderr := os.Stderr
	os.Stderr = w
	multi := io.MultiWriter(stderr, writer)
	reader := bufio.NewReader(r)
	go func() {
		for {
			_, err := io.Copy(multi, reader)
			if err != nil {
				panic("copying pipe reader to multi writer: " + err.Error())
			}
		}
	}()
	return nil
}
