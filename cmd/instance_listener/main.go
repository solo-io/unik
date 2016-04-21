package main

import (
	"net"
	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"time"
	"strings"
	"sync"
	"encoding/json"
	"net/http"
	"flag"
	"errors"
	"fmt"
)

func main() {
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	lock := sync.RWMutex{}
	macIpMap := make(map[string]string)

	listenerIp, err := getLocalIp()
	if err != nil {
		logrus.Fatalf("failed to get local ip: %v", err)
	}

	logrus.Infof("Starting unik discovery (udp heartbeat broadcast) with ip %s", listenerIp)
	info := []byte("unik:" + listenerIp)
	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: 9876,
	})
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"broadcast-ip": BROADCAST_IPv4,
		}).Fatalf("failed to dial udp broadcast connection")
	}
	go func() {
		for {
			_, err = socket.Write(info)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"broadcast-ip": BROADCAST_IPv4,
				}).Fatalf("failed writing to broadcast udp socket")
			}
			logrus.Debugf("broadcasting...")
			time.Sleep(2000 * time.Millisecond)
		}
	}()
	m := martini.Classic()
	m.Post("/register", func(res http.ResponseWriter, req *http.Request) {
		splitAddr := strings.Split(req.RemoteAddr, ":")
		if len(splitAddr) < 1 {
			logrus.WithFields(logrus.Fields{
				"req.RemoteAddr": req.RemoteAddr,
			}).Errorf("could not parse remote addr into ip/port combination")
			return
		}
		instanceIp := splitAddr[0]
		macAddress := req.URL.Query().Get("mac_address")
		logrus.WithFields(logrus.Fields{
			"Ip": instanceIp,
			"mac-address": macAddress,
		}).Infof("Instance registered")
		//mac address = the instance id in vsphere
		lock.Lock()
		defer lock.Unlock()
		macIpMap[macAddress] = instanceIp
	})
	m.Get("/instances", func() (string, error) {
		lock.RLock()
		defer lock.RUnlock()
		data, err := json.Marshal(macIpMap)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})
	m.RunOnAddr(":3000")
}

func getLocalIp() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", errors.New("retrieving network interfaces" + err.Error())
	}
	for _, iface := range ifaces {
		logrus.Infof("found an interface: %v\n", iface)
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				logrus.WithField("addr", addr).Debugf("inspecting address")
				switch v := addr.(type) {
				case *net.IPNet:
					if !v.IP.IsLoopback() && v.IP.IsGlobalUnicast() && v.IP.To4() != nil {
						return v.IP.To4().String(), nil
					}
				case *net.IPAddr:
					if !v.IP.IsLoopback() && v.IP.IsGlobalUnicast() && v.IP.To4() != nil {
						return v.IP.To4().String(), nil
					}
				}
			}
	}
	return "", errors.New("failed to find ip on ifaces: "+fmt.Sprintf("%v", ifaces))
}