package firecracker

import "github.com/emc-advanced-dev/pkg/errors"

func (p *FirecrackerProvider) DetachVolume(id string) error {

	return errors.New("not supported for firecracker", nil)
}
