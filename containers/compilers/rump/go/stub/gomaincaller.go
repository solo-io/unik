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
func gomaincaller(kludge_dns_addr C.uint, argc C.int, argv unsafe.Pointer) {
	os.Args = nil
	argcint := int(argc)
	argvarr := ((*[1 << 30]*C.char)(argv))
	for i := 0; i < argcint; i += 1 {
		os.Args = append(os.Args, C.GoString(argvarr[i]))
	}
	if err := stub(uint32(kludge_dns_addr)); err != nil {
		log.Printf("fatal: stub failed: %v", err)
		return
	}
	main()
}

const BROADCAST_LISTENING_PORT = 9967

func stub(dnsAddr uint32) error {
	//load dns information from c kludge_dns_addr
	b1 := (dnsAddr >> 0 * 8) & 0xFF
	b2 := (dnsAddr >> 1 * 8) & 0xFF
	b3 := (dnsAddr >> 2 * 8) & 0xFF
	b4 := (dnsAddr >> 3 * 8) & 0xFF
	nameserverString := fmt.Sprintf("nameserver %d.%d.%d.%d", b1, b2, b3, b4)

	log.Printf("writing dns addr: %s", nameserverString)

	if err := ioutil.WriteFile("/etc/resolv.conf", []byte(nameserverString), 0644); err != nil {
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
