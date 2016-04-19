package common

import (
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/layer-x/layerx-commons/lxerrors"
	"encoding/json"
	"fmt"
)

func GetInstanceIp(listenerIp string, listenerPort int, instanceId string) (string, error) {
	_, body, err := lxhttpclient.Get(fmt.Sprintf("%s:%v", listenerIp, listenerPort), "/instances", nil)
	if err != nil {
		return "", lxerrors.New("http GET on instance listener", err)
	}
	var instanceIpMap map[string]string
	if err := json.Unmarshal(body, &instanceIpMap); err != nil {
		return "", lxerrors.New("unmarshalling response ("+string(body)+") to map", err)
	}
	ip, ok := instanceIpMap[instanceId]
	if !ok {
		return "", lxerrors.New("instance not found in map", err)
	}
	return ip, nil
}