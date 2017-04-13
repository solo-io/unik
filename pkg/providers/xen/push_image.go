package xen

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/providers/common"
	"github.com/cf-unik/unik/pkg/types"
)

func (p *XenProvider) PushImage(params types.PushImagePararms) error {
	image, err := p.GetImage(params.ImageName)
	if err != nil {
		return errors.New("finding image for "+params.ImageName, err)
	}
	if err := common.PushImage(params.Config, image, getImagePath(image.Name)); err != nil {
		return errors.New("pushing image "+image.Name, err)
	}
	logrus.Infof("pushed image %v to %v", image.Name, params.Config.URL)
	return nil
}
