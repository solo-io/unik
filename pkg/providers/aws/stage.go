package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxlog"
	"os"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
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

func (p *AwsProvider) Stage(logger lxlog.Logger, name string, rawImage *types.RawImage, force bool) (_ *types.Image, err error) {
	var snapshotId, volumeId string
	s3svc := p.newS3(logger)
	ec2svc := p.newEC2(logger)
	defer func() {
		if err != nil {
			logger.WithErr(err).Errorf("aws staging encountered an error")
			if snapshotId != "" {
				logger.Warnf("cleaning up snapshot %s", snapshotId)
				deleteSnapshot(ec2svc, snapshotId)
			}
			if volumeId != "" {
				logger.Warnf("cleaning up volume %s", volumeId)
				deleteVolume(ec2svc, volumeId)
			}
		}
	}()

	defer func() {
		logger.Debugf("cleaninng up image %s", rawImage.LocalImagePath)
		os.Remove(rawImage.LocalImagePath)
	}()

	volumeId, err = createDataVolumeFromRawImage(s3svc, ec2svc, rawImage.LocalImagePath, p.config.Zone)
	if err != nil {
		return nil, lxerrors.New("creating aws boot volume", err)
	}

	createSnasphotInput := &ec2.CreateSnapshotInput{
		Description: aws.String("snapshot for unikernel image " + name),
		VolumeId: aws.String(volumeId),
	}
	createSnapshotOutput, err := ec2svc.CreateSnapshot(createSnasphotInput)
	if err != nil {
		return nil, lxerrors.New("creating aws snapshot", err)
	}
	snapshotId = *createSnapshotOutput.SnapshotId

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

	logger.WithFields(lxlog.Fields{
		"name": name,
		"architecture": architecture,
		"virtualization-type": virtualizationType,
		"kernel-id": kernelId,
		"block-device-mappings": blockDeviceMappings,
		"root-device-name": rootDeviceName,
	}).Infof("creating AMI for unikernel image")

	registerImageInput := &ec2.RegisterImageInput{
		Name: aws.String(name),
		Architecture:        aws.String(architecture),
		BlockDeviceMappings: blockDeviceMappings,
		RootDeviceName: aws.String(rootDeviceName),
		VirtualizationType:  aws.String(virtualizationType),
		KernelId:            aws.String(kernelId),
	}

	registerImageOutput, err := ec2svc.RegisterImage(registerImageInput)
	if err != nil {
		return nil, lxerrors.New("registering snapshot as image", err)
	}

	rawImageFile, err := os.Stat(rawImage.LocalImagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	image := &types.Image{
		Id: *registerImageOutput.ImageId,
		Name: name,
		DeviceMappings: rawImage.DeviceMappings,
		SizeMb: sizeMb,
		Infrastructure: types.Infrastructure_AWS,
		Created: time.Now(),
	}
	logger.WithFields(lxlog.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
