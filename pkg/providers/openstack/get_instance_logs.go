package openstack

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *OpenstackProvider) GetInstanceLogs(id string) (string, error) {
	return "", errors.New("not yet supportded for openstack", nil)
}
