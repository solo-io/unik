package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) ListVolumes() ([]*types.Volume, error) {
	if len(p.state.GetVolumes()) < 1 {
		return []*types.Volume{}, nil
	}
	volumeIds := []*string{}
	for volumeId := range p.state.GetVolumes() {
		volumeIds = append(volumeIds, aws.String(volumeId))
	}
	param := &ec2.DescribeVolumesInput{
		VolumeIds: volumeIds,
	}
	output, err := p.newEC2().DescribeVolumes(param)
	if err != nil {
		return nil, errors.New("running ec2 describe volumes ", err)
	}
	volumes := []*types.Volume{}
	for _, ec2Volume := range output.Volumes {
		volumeId := *ec2Volume.VolumeId
		if volumeId == "" {
			continue
		}
		volume, ok := p.state.GetVolumes()[volumeId]
		if !ok {
			logrus.WithFields(logrus.Fields{"ec2Volume": ec2Volume}).Errorf("found a volume that unik has no record of")
			continue
		}
		if len(ec2Volume.Attachments) > 0 {
			if len(ec2Volume.Attachments) > 1 {
				return nil, errors.New("ec2 reports volume to have >1 attachments. wut", nil)
			}
			volume.Attachment = *ec2Volume.Attachments[0].InstanceId
		} else {
			volume.Attachment = ""
		}
		err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
			volumes[volume.Id] = volume
			return nil
		})
		if err != nil {
			return nil, errors.New("modifying volume map in state", err)
		}
		err = p.state.Save()
		if err != nil {
			return nil, errors.New("saving modified volume map to state", err)
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}
