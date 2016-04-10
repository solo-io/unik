package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *AwsProvider) DeleteVolume(logger lxlog.Logger, id string, force bool) error {
	volume, err := p.GetVolume(logger, id)
	if err != nil {
		return lxerrors.New("retrieving volume "+id, err)
	}
	param := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volume.Id),
	}
	_, err = p.newEC2(logger).DeleteVolume(param)
	if err != nil {
		return lxerrors.New("failed to terminate volume "+volume.Id, err)
	}
	return p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})
}
