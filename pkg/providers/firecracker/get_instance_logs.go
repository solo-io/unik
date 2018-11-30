package firecracker

import "github.com/emc-advanced-dev/pkg/errors"

func (p *FirecrackerProvider) GetInstanceLogs(id string) (string, error) {
	return "", errors.New("not supported", nil)
}
