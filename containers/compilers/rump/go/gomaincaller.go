package main

import (
	"C"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const BROADCAST_LISTENING_PORT=9876
const EnvFile = "env.json"

var timeout = time.Duration(2 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func getEnvAmazon() (map[string]string, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: dialTimeout,
		},
	}
	resp, err := client.Get("http://169.254.169.254/latest/user-data")
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var env map[string]string
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}

func getEnvFile() (map[string]string, error) {
	data, err := ioutil.ReadFile(EnvFile)
	if err != nil {
		return nil, err
	}
	var env map[string]string
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}

func getEnvFromInject(req *http.Request) (map[string]string, error) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	var env map[string]string
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}

//export gomaincaller
func gomaincaller() {
	//make logs available via http request
	logs := bytes.Buffer{}
	if err := teeStdout(&logs); err != nil {
		log.Fatal(err)
	}
	if err := teeStderr(&logs); err != nil {
		log.Fatal(err)
	}

	log.Printf("unik boostrapping beginning...")

	go func() {
		listenerIp, err := getListenerIp()
		if err != nil {
			log.Printf("err getting listener ip: %v", err)
			return
		}
		if err := registerWithListener(listenerIp); err != nil {
			log.Printf("err registering with listener: %v", err)
			return
		}
	}()

	envChan := make(chan map[string]string)
	mux := http.NewServeMux()
	//serve logs
	mux.HandleFunc("/logs", func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(res, "logs: %s", string(logs.Bytes()))
	})
	//listen for injectable logs
	mux.HandleFunc("/inject_env", func(res http.ResponseWriter, req *http.Request) {
		log.Printf("someone injected my env: %v", req)
		env, _ := getEnvFromInject(req)
		envChan <- env
		fmt.Fprintf(res, "accepted")
	})
	log.Printf("starting log server\n")
	go http.ListenAndServe(fmt.Sprintf(":%v", BROADCAST_LISTENING_PORT), mux)

	errChan := make(chan error)
	go func() {
		env, err := getEnvFile()
		envChan <- env
		errChan <- err
	}()
	go func() {
		env, err := getEnvAmazon()
		envChan <- env
		errChan <- err
	}()

	errCounter := 0
	envLoop:
	for {
		log.Printf("waiting for env")
		select {
		case env := <-envChan:
			if env != nil {
				log.Printf("env was set: %v", env)
				setEnv(env)
				break envLoop
			}
		case err := <-errChan:
			if errCounter < 2 {
				log.Printf("error: %v", err)
				errCounter++
			} else {
				//all getEnv failed
				log.Fatal("err: %v", err)
			}
		}
	}

	log.Printf("calling main\n")
	main()
}

func setEnv(env map[string]string) error {
	for key, val := range env {
		os.Setenv(key, val)
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(EnvFile, data, 0644); err != nil {
		return err
	}
	return nil
}

func registerWithListener(listenerIp string) error {
	//get MAC Addr
	ifaces, err := net.Interfaces()
	if err != nil {
		return errors.New("retrieving network interfaces" + err.Error())
	}
	macAddress := ""
	for _, iface := range ifaces {
		log.Printf("found an interface: %v\n", iface)
		if len(iface.HardwareAddr) > 0 {
			macAddress = iface.HardwareAddr.String()
			break
		}
	}
	if macAddress == "" {
		return errors.New("could not find mac address")
	}

	if _, err := http.Post("http://" + listenerIp + ":3000/register?mac_address=" + macAddress, "", bytes.NewBuffer([]byte{})); err != nil {
		return err
	}
	return nil
}

func getListenerIp() (string, error) {
	log.Printf("listening for udp heartbeat...")
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9876,
	})
	if err != nil {
		return "", err
	}
	for {
		log.Printf("listening...")
		data := make([]byte, 4096)
		_, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			return "", err
		}
		log.Printf("recieved an ip from %s with data: %s", remoteAddr.IP.String(), string(data))
		if strings.Contains(string(data), "unik") {
			data = bytes.Trim(data, "\x00")
			return strings.Split(string(data), ":")[1], nil
		}
	}
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
				log.Fatalf("copying pipe reader to multi writer: " + err.Error())
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
	stdout := os.Stderr
	os.Stderr = w
	multi := io.MultiWriter(stdout, writer)
	reader := bufio.NewReader(r)
	go func() {
		for {
			_, err := io.Copy(multi, reader)
			if err != nil {
				log.Fatalf("copying pipe reader to multi writer: " + err.Error())
			}
		}
	}()
	return nil
}
