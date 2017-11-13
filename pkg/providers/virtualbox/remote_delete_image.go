package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/types"
)

func (p *VirtualboxProvider) RemoteDeleteImage(params types.RemoteDeleteImagePararms) error {
	if err := common.RemoteDeleteImage(params.Config, getImagePath(params.ImageName)); err != nil {
		return errors.New("deleting image "+params.ImageName, err)
	}
	logrus.Infof("pushed image %v to %v", params.ImageName, params.Config.URL)
	return nil
}
