package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) DeleteVolume(id string, force bool) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return errors.New("retrieving volume "+id, err)
	}
	if volume.Attachment != "" {
		if force {
			if err := p.DetachVolume(volume.Id); err != nil {
				return errors.New("detaching volume for deletion", err)
			} else {
				return errors.New("volume "+volume.Id+" is attached to instance."+volume.Attachment+", try again with --force or detach volume first", err)
			}
		}
	}
	param := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volume.Id),
	}
	_, err = p.newEC2().DeleteVolume(param)
	if err != nil {
		return errors.New("failed to terminate volume "+volume.Id, err)
	}
	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		delete(volumes, volume.Id)
		return nil
	})
	if err != nil {
		return errors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return errors.New("saving volume map to state", err)
	}
	return nil
}
