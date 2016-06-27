package photon

import "github.com/emc-advanced-dev/pkg/errors"

func (p *PhotonProvider) GetInstanceLogs(id string) (string, error) {
	return "", errors.New("not supported", nil)
}
