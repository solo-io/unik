package main

import (
	"C"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"unsafe"
)

//export gomaincaller
func gomaincaller(argc C.int, argv unsafe.Pointer) {
	os.Args = nil
	argcint := int(argc)
	argvarr := ((*[1 << 30]*C.char)(argv))
	for i := 0; i < argcint; i += 1 {
		os.Args = append(os.Args, C.GoString(argvarr[i]))
	}
	if err := stub(); err != nil {
		log.Printf("fatal: stub failed: %v", err)
		return
	}
	main()
}

const BROADCAST_LISTENING_PORT = 9967

func stub() error {
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
