// +build gcloud

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
	ioutil.WriteFile()

	log.Printf("bootstrapping using gcloud metadata service")
	env, err := getEnvGcloud()
	if err != nil {
		return errors.New("failed to get env from gcloud: " + err.Error())
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

func getEnvGcloud() (map[string]string, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: dialTimeout,
		},
	}
	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/attributes/ENV_DATA", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Metadata-Flavor", "Google")
	resp, err := client.Do(req)
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
