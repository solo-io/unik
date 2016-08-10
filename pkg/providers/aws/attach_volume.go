package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) AttachVolume(id, instanceId, mntPoint string) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return errors.New("retrieving volume "+id, err)
	}
	if volume.Attachment != "" {
		return errors.New("volume is already attached to instance "+volume.Attachment, nil)
	}
	instance, err := p.GetInstance(instanceId)
	if err != nil {
		return errors.New("retrieving instance "+instanceId, err)
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return errors.New("retrieving image for instance", err)
	}
	if err := common.VerifyMntsInput(p, image, map[string]string{mntPoint: id}); err != nil {
		return errors.New("invalid mapping for volume", err)
	}
	deviceName, err := common.GetDeviceNameForMnt(image, mntPoint)
	if err != nil {
		logrus.WithFields(logrus.Fields{"image": image.Id, "mappings": image.RunSpec.DeviceMappings, "mount point": mntPoint}).Errorf("given mapping was not found for image")
		return err
	}
	param := &ec2.AttachVolumeInput{
		VolumeId:   aws.String(volume.Id),
		InstanceId: aws.String(instance.Id),
		Device:     aws.String(deviceName),
	}
	_, err = p.newEC2().AttachVolume(param)
	if err != nil {
		return errors.New("failed to attach volume "+volume.Id, err)
	}
	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[volume.Id]
		if !ok {
			return errors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = instance.Id
		return nil
	})
	if err != nil {
		return errors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return errors.New("saving volume to state", err)
	}
	return nil
}
