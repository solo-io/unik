package ukvm

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
)

func (p *UkvmProvider) PushImage(params types.PushImagePararms) error {
	return errors.New("pushing image not supported for ukvm", nil)
}
