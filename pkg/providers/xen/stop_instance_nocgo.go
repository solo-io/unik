// +build !cgo

package qemu

import "github.com/emc-advanced-dev/pkg/errors"

func (p *XenProvider) StopInstance(id string) error {

	return errors.New("Stopping qemu instance is not supported without cgo", nil)
}
