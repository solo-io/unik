package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/Sirupsen/logrus"
)

const UNIK_VOLUME_ID = "UNIK_VOLUME_ID"

func (p *AwsProvider) ListVolumes() ([]*types.Volume, error) {
	param := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("tag-key"),
				Values: []*string{aws.String(UNIK_VOLUME_ID)},
			},
		},
	}
	output, err := p.newEC2().DescribeVolumes(param)
	if err != nil {
		return nil, lxerrors.New("running ec2 describe volumes ", err)
	}
	volumes := []*types.Volume{}
	for _, ec2Volume := range output.Volumes {
		volumeId := parseVolumeId(ec2Volume)
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
				return nil, lxerrors.New("ec2 reports volume to have >1 attachments. wut", nil)
			}
			volume.Attachment = *ec2Volume.Attachments[0].InstanceId
		} else {
			volume.Attachment = ""
		}
		p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
			volumes[volume.Id] = volume
			return nil
		})
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

func parseVolumeId(ec2Volume *ec2.Volume) string {
	for _, tag := range ec2Volume.Tags {
		if *tag.Key == UNIK_VOLUME_ID {
			return *tag.Value
		}
	}
	return ""
}
