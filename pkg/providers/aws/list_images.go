package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

const UNIK_IMAGE_ID = "UNIK_IMAGE_ID"

func (p *AwsProvider) ListImages(logger lxlog.Logger) ([]*types.Image, error) {
	param := &ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("tag-key"),
				Values: []*string{aws.String(UNIK_IMAGE_ID)},
			},
		},
	}
	output, err := p.newEC2(logger).DescribeImages(param)
	if err != nil {
		return nil, lxerrors.New("running ec2 describe images ", err)
	}
	images := []*types.Image{}
	for _, ec2Image := range output.Images {
		imageId := parseImageId(ec2Image)
		if imageId == "" {
			continue
		}
		image, ok := p.State.GetImages()[imageId]
		if !ok {
			logger.WithFields(lxlog.Fields{"ec2Image": ec2Image}).Errorf("found an image that unik has no record of")
			continue
		}
		images = append(images, image)
	}
	return images, nil
}

func parseImageId(ec2Image *ec2.Image) string {
	for _, tag := range ec2Image.Tags {
		if *tag.Key == UNIK_IMAGE_ID {
			return *tag.Value
		}
	}
	return ""
}
