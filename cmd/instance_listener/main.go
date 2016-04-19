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
)

func main() {
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	lock := sync.RWMutex{}
	macIpMap := make(map[string]string)

	logrus.Infof("Starting unik discovery (udp heartbeat broadcast)")
	info := []byte("unik")
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
	go func(){
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