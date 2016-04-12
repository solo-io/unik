package common

import (
	"fmt"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/Sirupsen/logrus"
)

const UnikLogsPort = 9876

func GetInstanceLogs(instance *types.Instance) (string, error) {
	if instance.IpAddress == "" {
		return "", lxerrors.New("instance has not been assigned a public ip address", nil)
	}
	_, body, err := lxhttpclient.Get(instance.IpAddress+fmt.Sprintf(":%v", UnikLogsPort), "/logs", nil)
	if err != nil {
		return "", lxerrors.New("faiiled to connect to instance at "+instance.IpAddress+" for logs", err)
	}
	logrus.WithFields(logrus.Fields{"response-length": len(body), "instance": instance}).Debugf("received stdout from instance")
	return string(body), nil
}
