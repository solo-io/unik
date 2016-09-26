package ukvm

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *UkvmProvider) PullImage(params types.PullImagePararms) error {

	return errors.New("pulling image not supported for ukvm", nil)
}
