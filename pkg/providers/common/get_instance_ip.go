package common

import (
	"encoding/json"
	"fmt"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
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
		return "", lxerrors.New("instance "+instanceId+" not found in map: "+fmt.Sprintf("%v", instanceIpMap), err)
	}
	return ip, nil
}
