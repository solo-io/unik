package aws

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) DetachVolume(logger lxlog.Logger, id string) error {
	volume, err := p.GetVolume(logger, id)
	if err != nil {
		return lxerrors.New("retrieving volume "+id, err)
	}
	param := &ec2.DetachVolumeInput{
		VolumeId: aws.String(volume.Id),
		Force: aws.Bool(true),
	}
	_, err = p.newEC2(logger).DetachVolume(param)
	if err != nil {
		return lxerrors.New("failed to detach volume "+volume.Id, err)
	}
	return p.State.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[volume.Id]
		if !ok {
			return lxerrors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = ""
		return nil
	})
}