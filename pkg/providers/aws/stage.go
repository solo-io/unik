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

var kernelIdMap = map[string]string{
	"ap-northeast-1": "aki-176bf516",
	"ap-southeast-1": "aki-503e7402",
	"ap-southeast-2": "aki-c362fff9",
	"eu-central-1":   "aki-184c7a05",
	"eu-west-1":      "aki-52a34525",
	"sa-east-1":      "aki-5553f448",
	"us-east-1":      "aki-919dcaf8",
	"us-gov-west-1":  "aki-1de98d3e",
	"us-west-1":      "aki-880531cd",
	"us-west-2":      "aki-fc8f11cc",
}

func (p *AwsProvider) Stage(name string, rawImage *types.RawImage, force bool) (_ *types.Image, err error) {
	images, err := p.ListImages()
	if err != nil {
		return nil, lxerrors.New("retrieving image list for existing image", err)
	}
	for _, image := range images {
		if image.Name == name {
			if !force {
				return nil, lxerrors.New("an image already exists with name '"+name+"', try again with --force", nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name " + name)
				err = p.DeleteImage(image.Id, true)
				if err != nil {
					return nil, lxerrors.New("removing previously existing image", err)
				}
			}
		}
	}

	var snapshotId, volumeId string
	s3svc := p.newS3()
	ec2svc := p.newEC2()
	defer func() {
		if err != nil {
			logrus.WithError(err).Errorf("aws staging encountered an error")
			if snapshotId != "" {
				logrus.Warnf("cleaning up snapshot %s", snapshotId)
				deleteSnapshot(ec2svc, snapshotId)
			}
			if volumeId != "" {
				logrus.Warnf("cleaning up volume %s", volumeId)
				deleteVolume(ec2svc, volumeId)
			}
		}
	}()

	logrus.WithField("raw-image", rawImage).WithField("az", p.config.Zone).Infof("creating boot volume from raw image")
	volumeId, err = createDataVolumeFromRawImage(s3svc, ec2svc, rawImage.LocalImagePath, p.config.Zone)
	if err != nil {
		return nil, lxerrors.New("creating aws boot volume", err)
	}

	logrus.WithField("volume-id", volumeId).Infof("creating snapshot from boot volume")
	createSnasphotInput := &ec2.CreateSnapshotInput{
		Description: aws.String("snapshot for unikernel image " + name),
		VolumeId:    aws.String(volumeId),
	}
	createSnapshotOutput, err := ec2svc.CreateSnapshot(createSnasphotInput)
	if err != nil {
		return nil, lxerrors.New("creating aws snapshot", err)
	}
	snapshotId = *createSnapshotOutput.SnapshotId

	snapDesc := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(snapshotId)},
	}
	err = ec2svc.WaitUntilSnapshotCompleted(snapDesc)
	if err != nil {
		return nil, lxerrors.New("waiting for snapshot to complete", err)
	}

	blockDeviceMappings := []*ec2.BlockDeviceMapping{}
	rootDeviceName := ""
	for _, deviceMapping := range rawImage.DeviceMappings {
		if deviceMapping.MountPoint == "/" {
			blockDeviceMappings = append(blockDeviceMappings, &ec2.BlockDeviceMapping{
				DeviceName: aws.String(deviceMapping.DeviceName),
				Ebs: &ec2.EbsBlockDevice{
					SnapshotId: aws.String(snapshotId),
				},
			})
			rootDeviceName = deviceMapping.DeviceName
			break
		}
	}
	if len(blockDeviceMappings) < 1 {
		return nil, lxerrors.New("did not find root device mapping for image", nil)
	}

	architecture := "x86_64"
	virtualizationType := "paravirtual"
	kernelId := kernelIdMap[p.config.Region]

	logrus.WithFields(logrus.Fields{
		"name":                  name,
		"architecture":          architecture,
		"virtualization-type":   virtualizationType,
		"kernel-id":             kernelId,
		"block-device-mappings": blockDeviceMappings,
		"root-device-name":      rootDeviceName,
	}).Infof("creating AMI for unikernel image")

	registerImageInput := &ec2.RegisterImageInput{
		Name:                aws.String(name),
		Architecture:        aws.String(architecture),
		BlockDeviceMappings: blockDeviceMappings,
		RootDeviceName:      aws.String(rootDeviceName),
		VirtualizationType:  aws.String(virtualizationType),
		KernelId:            aws.String(kernelId),
	}

	registerImageOutput, err := ec2svc.RegisterImage(registerImageInput)
	if err != nil {
		return nil, lxerrors.New("registering snapshot as image", err)
	}

	imageId := *registerImageOutput.ImageId

	logrus.WithField("volume-id", volumeId).Infof("tagging image, snapshot, and volume with unikernel id")
	tagObjects := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(imageId),
			aws.String(snapshotId),
			aws.String(volumeId),
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String(UNIK_IMAGE_ID),
				Value: aws.String(imageId),
			},
			&ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}
	_, err = ec2svc.CreateTags(tagObjects)
	if err != nil {
		return nil, lxerrors.New("tagging snapshot, image, and volume", err)
	}

	rawImageFile, err := os.Stat(rawImage.LocalImagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	image := &types.Image{
		Id:             imageId,
		Name:           name,
		DeviceMappings: rawImage.DeviceMappings,
		SizeMb:         sizeMb,
		Infrastructure: types.Infrastructure_AWS,
		Created:        time.Now(),
	}

	err = p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[imageId] = image
		return nil
	})
	if err != nil {
		return nil, lxerrors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, lxerrors.New("saving image map to state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
