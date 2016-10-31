// +build ec2

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func bootstrap() error {
	log.Printf("bootstrapping using ec2 metadata service")
	env, err := getEnvAmazon()
	if err != nil {
		return errors.New("failed to get env from ec2: " + err.Error())
	}
	if err := setEnv(env); err != nil {
		return errors.New("setting env: " + err.Error())
	}
	return nil
}

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
