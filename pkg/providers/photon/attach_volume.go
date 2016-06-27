package photon

import "github.com/emc-advanced-dev/pkg/errors"

func (p *PhotonProvider) AttachVolume(id, instanceId, mntPoint string) error {
	return errors.New("not yet supportded for photon", nil)
}
