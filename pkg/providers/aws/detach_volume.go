package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) DetachVolume(id string) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return errors.New("retrieving volume "+id, err)
	}
	if volume.Attachment == "" {
		return errors.New("volume has no attachment", nil)
	}
	param := &ec2.DetachVolumeInput{
		VolumeId: aws.String(volume.Id),
		Force:    aws.Bool(true),
	}
	_, err = p.newEC2().DetachVolume(param)
	if err != nil {
		return errors.New("failed to detach volume "+volume.Id, err)
	}
	if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[volume.Id]
		if !ok {
			return errors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = ""
		return nil
	}); err != nil {
		return errors.New("modifying volume map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return errors.New("saving modified volume map to state", err)
	}
	return nil
}
