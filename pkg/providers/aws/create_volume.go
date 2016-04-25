package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"time"
)

func (p *AwsProvider) CreateVolume(name, imagePath string) (*types.Volume, error) {
	logrus.WithField("raw-image", imagePath).WithField("az", p.config.Zone).Infof("creating data volume from raw image")
	s3svc := p.newS3()
	ec2svc := p.newEC2()
	volumeId, err := createDataVolumeFromRawImage(s3svc, ec2svc, imagePath, p.config.Zone)
	if err != nil {
		return nil, lxerrors.New("creating aws boot volume", err)
	}
	tagVolumeInput := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(volumeId),
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}
	_, err = ec2svc.CreateTags(tagVolumeInput)
	if err != nil {
		return nil, lxerrors.New("tagging volume", err)
	}

	rawImageFile, err := os.Stat(imagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	volume := &types.Volume{
		Id:             volumeId,
		Name:           name,
		SizeMb:         sizeMb,
		Attachment:     "",
		Infrastructure: types.Infrastructure_AWS,
		Created:        time.Now(),
	}

	err = p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volumes[volume.Id] = volume
		return nil
	})
	if err != nil {
		return nil, lxerrors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, lxerrors.New("saving volume map to state", err)
	}

	return nil, nil
}
func (p *AwsProvider) CreateEmptyVolume(name string, size int) (*types.Volume, error) {
	return nil, nil
}
