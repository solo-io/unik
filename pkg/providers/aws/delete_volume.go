package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *AwsProvider) DeleteVolume(id string, force bool) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return lxerrors.New("retrieving volume "+id, err)
	}
	param := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volume.Id),
	}
	_, err = p.newEC2().DeleteVolume(param)
	if err != nil {
		return lxerrors.New("failed to terminate volume "+volume.Id, err)
	}
	return p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})
}
