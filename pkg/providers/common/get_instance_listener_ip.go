package common

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"net"
	"strings"
	"time"
)

var socket *net.UDPConn

const BROADCAST_PORT = 9967

func GetInstanceListenerIp(dataPrefix string, timeout time.Duration) (string, error) {
	errc := make(chan error)
	go func() {
		<-time.After(timeout)
		errc <- errors.New("getting instance listener ip timed out after "+timeout.String(), nil)
	}()
	logrus.Infof("listening for udp heartbeat...")
	var err error
	//only initialize socket once
	logrus.Debug("ARE WE LISTENING ON THE SOCKET YET?", socket)
	if socket == nil {
		socket, err = net.ListenUDP("udp4", &net.UDPAddr{
			IP:   net.IPv4(0, 0, 0, 0),
			Port: BROADCAST_PORT,
		})
		logrus.Debug("socket was", socket, "err was", err)
		if err != nil {
			return "", errors.New("opening udp socket", err)
		}
	}
	resultc := make(chan string)
	var stopLoop bool
	go func() {
		logrus.Infof("UDP Server listening on %s:%v", "0.0.0.0", BROADCAST_PORT)
		for !stopLoop {
			data := make([]byte, 4096)
			_, remoteAddr, err := socket.ReadFromUDP(data)
			if err != nil {
				errc <- errors.New("reading udp data", err)
				return
			}
			logrus.Infof("received an ip from %s with data: %s", remoteAddr.IP.String(), string(data))
			if strings.Contains(string(data), dataPrefix) {
				data = bytes.Trim(data, "\x00")
				resultc <- strings.Split(string(data), ":")[1]
				return
			}
		}
	}()
	select {
	case result := <-resultc:
		return result, nil
	case err := <-errc:
		stopLoop = true
		return "", err
	}
}
