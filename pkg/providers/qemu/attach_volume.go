package qemu

import "github.com/emc-advanced-dev/pkg/errors"

func (p *QemuProvider) AttachVolume(id, instanceId, mntPoint string) error {
	return errors.New("not supported", nil)
}
