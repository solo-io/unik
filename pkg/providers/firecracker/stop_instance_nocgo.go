// +build !cgo

package firecracker

import "github.com/emc-advanced-dev/pkg/errors"

func (p *FirecrackerProvider) StopInstance(id string) error {

	return errors.New("Stopping firecracker instance is not supported without cgo", nil)
}
