// +build !cgo

package ukvm

import "github.com/emc-advanced-dev/pkg/errors"

func (p *UkvmProvider) StopInstance(id string) error {

	return errors.New("Stopping ukvm instance is not supported without cgo", nil)
}
