package common

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"strings"
	"net"
	"github.com/Sirupsen/logrus"
	"bytes"
	"time"
)

func GetInstanceListenerIp(timeout time.Duration) (string, error) {
	closeChan := make(chan struct{})
	go func(){
		<-time.After(timeout)
		close(closeChan)
	}()
	logrus.Infof("listening for udp heartbeat...")
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9876,
	})
	if err != nil {
		return "", errors.New("opening udp socket", err)
	}
	resultc := make(chan string)
	errc := make(chan error)
	go func(){
		logrus.Infof("UDP Server listening on %s:%v", "0.0.0.0", 9876)
		data := make([]byte, 4096)
		_, remoteAddr, err := socket.ReadFromUDP(data)
		if err != nil {
			errc <- errors.New("reading udp data", err)
		}
		logrus.Infof("received an ip from %s with data: %s", remoteAddr.IP.String(), string(data))
		if strings.Contains(string(data), "unik") {
			data = bytes.Trim(data, "\x00")
			resultc <- strings.Split(string(data), ":")[1]
		}
	}()
	select {
	case <-closeChan:
		return "", errors.New("getting instance listener ip timed out after "+timeout.String(), nil)
	case result := <- resultc:
		return result, nil
	case err := <- errc:
		return "", err
	}
}