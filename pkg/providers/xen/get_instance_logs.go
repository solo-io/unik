package qemu

import "github.com/emc-advanced-dev/pkg/errors"

func (p *XenProvider) GetInstanceLogs(id string) (string, error) {
	return "", errors.New("not supported", nil)
}
