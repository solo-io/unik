package qemu

import "github.com/emc-advanced-dev/pkg/errors"

func (p *QemuProvider) DetachVolume(id string) error {
	return errors.New("not supported", nil)
}
