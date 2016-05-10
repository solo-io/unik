package qemu

import "github.com/emc-advanced-dev/pkg/errors"

func (p *QemuProvider) DeleteVolume(id string, force bool) error {
	return errors.New("not supported", nil)
}
