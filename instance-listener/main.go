package main

//TODO: make sure to always binpack this file to bindata/instance_listener_data on recompile
//add a make target

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const statefile = "/data/statefile.json"

type state struct {
	MacIpMap  map[string]string            `json:"Ips"`
	MacEnvMap map[string]map[string]string `json:"Envs"`
}

func main() {
	args := os.Args
	for i, arg := range args {
		log.Printf("arg %v: %s", i, arg)
	}
	dataPrefix := flag.String("prefix", "unik_", "prefix for data sent via udp (for identification purposes")
	flag.Parse()
	for i, arg := range flag.Args() {

		log.Printf("flagarg %v: %s", i, arg)
	}
	if *dataPrefix == "unik_" {
		log.Printf("ERROR: must provide -prefix")
		return
	}
	if *dataPrefix == "" {
		log.Printf("ERROR: -prefix cannot be \"\"")
		return
	}
	ipMapLock := sync.RWMutex{}
	envMapLock := sync.RWMutex{}
	saveLock := sync.Mutex{}
	var s state
	s.MacIpMap = make(map[string]string)
	s.MacEnvMap = make(map[string]map[string]string)

	data, err := ioutil.ReadFile(statefile)
	if err != nil {
		log.Printf("could not read statefile, maybe this is first boot: " + err.Error())
	} else {
		if err := json.Unmarshal(data, &s); err != nil {
			log.Printf("failed to parse state json: " + err.Error())
		}
	}

	listenerIp, err := getLocalIp()
	if err != nil {
		log.Printf("ERROR: failed to get local ip: %v", err)
		return
	}

	log.Printf("Starting unik discovery (udp heartbeat broadcast) with ip %s", listenerIp.String())
	info := []byte(*dataPrefix + ":" + listenerIp.String())
	listenerIpMask := listenerIp.DefaultMask()
	BROADCAST_IPv4 := reverseMask(listenerIp, listenerIpMask)
	if listenerIpMask == nil {
		log.Printf("ERROR: listener-ip: %v; listener-ip-mask: %v; could not calculate broadcast address", listenerIp, listenerIpMask)
		return
	}
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: 9876,
	})
	if err != nil {
		log.Printf(fmt.Sprintf("ERROR: broadcast-ip: %v; failed to dial udp broadcast connection", BROADCAST_IPv4))
		return
	}
	go func() {
		log.Printf("broadcasting...")
		for {
			_, err = socket.Write(info)
			if err != nil {
				log.Printf("ERROR: broadcast-ip: %v; failed writing to broadcast udp socket: "+err.Error(), BROADCAST_IPv4)
				return
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	m := http.NewServeMux()
	m.HandleFunc("/register", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		splitAddr := strings.Split(req.RemoteAddr, ":")
		if len(splitAddr) < 1 {
			log.Printf("req.RemoteAddr: %v, could not parse remote addr into ip/port combination", req.RemoteAddr)
			return
		}
		instanceIp := splitAddr[0]
		macAddress := req.URL.Query().Get("mac_address")
		log.Printf("Instance registered")
		log.Printf("ip: %v", instanceIp)
		log.Printf("ip: %v", macAddress)
		//mac address = the instance id in vsphere/vbox
		go func() {
			ipMapLock.Lock()
			defer ipMapLock.Unlock()
			s.MacIpMap[macAddress] = instanceIp
			go save(s, saveLock)
		}()
		envMapLock.RLock()
		defer envMapLock.RUnlock()
		env, ok := s.MacEnvMap[macAddress]
		if !ok {
			env = make(map[string]string)
			log.Printf("mac: %v", macAddress)
			log.Printf("env: %v", s.MacEnvMap)
			log.Printf("no env set for instance, replying with empty map")
		}
		data, err := json.Marshal(env)
		if err != nil {
			log.Printf("could not marshal env to json: " + err.Error())
			return
		}
		log.Printf("responding with data: %s", data)
		fmt.Fprintf(res, "%s", data)
	})
	m.HandleFunc("/set_instance_env", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		macAddress := req.URL.Query().Get("mac_address")
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		defer req.Body.Close()
		var env map[string]string
		if err := json.Unmarshal(data, &env); err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
			return
		}
		log.Printf("Env set for instance")
		log.Printf("mac: %v", macAddress)
		log.Printf("env: %v", env)
		envMapLock.Lock()
		defer envMapLock.Unlock()
		s.MacEnvMap[macAddress] = env
		go save(s, saveLock)
		res.WriteHeader(http.StatusAccepted)
	})
	m.HandleFunc("/instances", func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			res.WriteHeader(http.StatusNotFound)
			return
		}
		ipMapLock.RLock()
		defer ipMapLock.RUnlock()
		data, err := json.Marshal(s.MacIpMap)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			res.Write([]byte(err.Error()))
		}
		res.Write(data)
	})
	log.Printf("listening on port 3000")
	http.ListenAndServe(":3000", m)
}

func getLocalIp() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return net.IP{}, errors.New("retrieving network interfaces" + err.Error())
	}
	for _, iface := range ifaces {
		log.Printf("found an interface: %v\n", iface)
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			log.Printf("inspecting address: %v", addr)
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() && v.IP.IsGlobalUnicast() && v.IP.To4() != nil {
					return v.IP.To4(), nil
				}
			}
		}
	}
	return net.IP{}, errors.New("failed to find ip on ifaces: " + fmt.Sprintf("%v", ifaces))
}

// ReverseMask returns the result of masking the IP address ip with mask.
func reverseMask(ip net.IP, mask net.IPMask) net.IP {
	n := len(ip)
	if n != len(mask) {
		return nil
	}
	out := make(net.IP, n)
	for i := 0; i < n; i++ {
		out[i] = ip[i] | (^mask[i])
	}
	return out
}

func save(s state, l sync.Mutex) {
	if err := func() error {
		l.Lock()
		defer l.Unlock()
		data, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(statefile, data, 0644); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		log.Printf("failed to save state file %s", statefile)
	}
}
