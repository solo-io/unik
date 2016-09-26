package ukvm

import "github.com/emc-advanced-dev/pkg/errors"

func (p *UkvmProvider) GetInstanceLogs(id string) (string, error) {
	return "", errors.New("not supported", nil)
}
