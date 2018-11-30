package aws

import (
	"github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/types"
	"os"
	"time"
)

func (p *AwsProvider) CreateVolume(params types.CreateVolumeParams) (*types.Volume, error) {
	logrus.WithField("raw-image", params.ImagePath).WithField("az", p.config.Zone).Infof("creating data volume from raw image")
	s3svc := p.newS3()
	ec2svc := p.newEC2()
	imageFile, err := os.Stat(params.ImagePath)
	if err != nil {
		return nil, errors.New("stat image file", err)
	}
	volumeId, err := createDataVolumeFromRawImage(s3svc, ec2svc, params.ImagePath, imageFile.Size(), types.ImageFormat_RAW, p.config.Zone)
	if err != nil {
		return nil, errors.New("creating aws boot volume", err)
	}
	tagVolumeInput := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(volumeId),
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(params.Name),
			},
		},
	}
	if _, err := ec2svc.CreateTags(tagVolumeInput); err != nil {
		return nil, errors.New("tagging volume", err)
	}

	rawImageFile, err := os.Stat(params.ImagePath)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	volume := &types.Volume{
		Id:             volumeId,
		Name:           params.Name,
		SizeMb:         sizeMb,
		Attachment:     "",
		Infrastructure: types.Infrastructure_AWS,
		Created:        time.Now(),
	}

	if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volumes[volume.Id] = volume
		return nil
	}); err != nil {
		return nil, errors.New("modifying volume map in state", err)
	}

	return nil, nil
}
func (p *AwsProvider) CreateEmptyVolume(name string, size int) (*types.Volume, error) {
	return nil, nil
}
