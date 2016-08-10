package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const UNIK_IMAGE_ID = "UNIK_IMAGE_ID"

func (p *AwsProvider) ListImages() ([]*types.Image, error) {
	if len(p.state.GetImages()) < 1 {
		return []*types.Image{}, nil
	}
	imageIds := []*string{}
	for imageId := range p.state.GetImages() {
		imageIds = append(imageIds, aws.String(imageId))
	}
	param := &ec2.DescribeImagesInput{
		ImageIds: imageIds,
	}
	output, err := p.newEC2().DescribeImages(param)
	if err != nil {
		return nil, errors.New("running ec2 describe images ", err)
	}
	images := []*types.Image{}
	for _, ec2Image := range output.Images {
		imageId := *ec2Image.ImageId
		image, ok := p.state.GetImages()[imageId]
		if !ok {
			logrus.WithFields(logrus.Fields{"ec2Image": ec2Image}).Errorf("found an image that unik has no record of")
			continue
		}
		images = append(images, image)
	}
	return images, nil
}
