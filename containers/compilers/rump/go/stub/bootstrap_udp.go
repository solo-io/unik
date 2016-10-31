// +build udp

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

func bootstrap() error {
	log.Printf("bootstrapping using instance listener on port %v", BROADCAST_LISTENING_PORT)
	listenerIp, err := getListenerIp()
	if err != nil {
		return errors.New("getting listener ip: " + err.Error())
	}
	env, err := registerWithListener(listenerIp)
	if err != nil {
		return errors.New("registering with listener: " + err.Error())
	}
	if err := setEnv(env); err != nil {
		return errors.New("setting env: " + err.Error())
	}
	return nil
}

func getListenerIp() (string, error) {
	log.Printf("listening for udp heartbeat...")
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9967,
	})
	if err != nil {
		return "", err
	}
	for {
		log.Printf("UDP Server listening on %s:%v", "0.0.0.0", 9967)
		data := make([]byte, 4096)
		_, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			return "", err
		}
		log.Printf("received an ip from %s with data: %s", remoteAddr.IP.String(), string(data))
		if strings.Contains(string(data), "unik") {
			data = bytes.Trim(data, "\x00")
			return strings.Split(string(data), ":")[1], nil
		}
	}
}

func registerWithListener(listenerIp string) (map[string]string, error) {
	//get MAC Addr
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, errors.New("retrieving network interfaces" + err.Error())
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
		return nil, errors.New("could not find mac address")
	}

	resp, err := http.Post("http://"+listenerIp+":3000/register?mac_address="+macAddress, "", bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var env map[string]string
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return env, nil
}
