package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"encoding/json"
	"encoding/base64"
	"time"
	"github.com/Sirupsen/logrus"
)

func (p *AwsProvider) RunInstance(name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
	"image-id": imageId,
		"mounts": mntPointsToVolumeIds,
		"env": env,
	}).Infof("running instance %s", name)

	var instanceId string
	ec2svc := p.newEC2()

	defer func(){
		if err != nil {
			logrus.WithError(err).Errorf("aws running instance encountered an error")
			if instanceId != "" {
				logrus.Warnf("cleaning up instance %s", instanceId)
				terminateInstanceInput := &ec2.TerminateInstancesInput{
					InstanceIds: []*string{aws.String(instanceId)},
				}
				ec2svc.TerminateInstances(terminateInstanceInput)
			}
		}
	}()

	image, err := p.GetImage(imageId)
	if err != nil {
		return nil, lxerrors.New("getting image", err)
	}
	err = verifyMntsInput(image, mntPointsToVolumeIds)
	if err != nil {
		return nil, err
	}

	envData, err := json.Marshal(env)
	if err != nil {
		return nil, lxerrors.New("could not convert instance env to json", err)
	}
	encodedData := base64.StdEncoding.EncodeToString(envData)
	runInstanceInput := &ec2.RunInstancesInput{
		ImageId: aws.String(image.Id),
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
		UserData: aws.String(encodedData),
	}

	runInstanceOutput, err := ec2svc.RunInstances(runInstanceInput)
	if err != nil {
		return nil, lxerrors.New("failed to run instance", err)
	}
	if len(runInstanceOutput.Instances) < 1 {
		logrus.WithFields(logrus.Fields{"output": runInstanceOutput}).Errorf("run instance %s failed, produced %v instances, expected 1", name, len(runInstanceOutput.Instances))
		return nil, lxerrors.New("expected 1 instance to be created", nil)
	}
	instanceId = *runInstanceOutput.Instances[0].InstanceId

	if len(runInstanceOutput.Instances) > 1 {
		logrus.WithFields(logrus.Fields{"output": runInstanceOutput}).Errorf("run instance %s failed, produced %v instances, expected 1", name, len(runInstanceOutput.Instances))
		return nil, lxerrors.New("expected 1 instance to be created", nil)
	}

	if len(mntPointsToVolumeIds) > 0 {
		logrus.Debugf("stopping instance for volume attach")
		err = p.StopInstance(instanceId)
		if err != nil {
			return nil, lxerrors.New("failed to stop instance for attachin volumes", err)
		}
		for mountPoint, volumeId := range mntPointsToVolumeIds {
			logrus.WithFields(logrus.Fields{"volume-id": volumeId}).Debugf("attaching volume %s to intance %s", volumeId, instanceId)
			err = p.AttachVolume(volumeId, instanceId, mountPoint)
			if err != nil {
				return nil, lxerrors.New("attaching volume to instance", err)
			}
		}
		err = p.StartInstance(instanceId)
		if err != nil {
			return nil, lxerrors.New("starting instance after volume attach", err)
		}
	}

	tagObjects := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(instanceId),
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:  aws.String(UNIK_INSTANCE_ID),
				Value: aws.String(instanceId),
			},
			&ec2.Tag{
				Key:  aws.String("Name"),
				Value: aws.String(name),
			},
		},
	}
	_, err = ec2svc.CreateTags(tagObjects)
	if err != nil {
		return nil, lxerrors.New("tagging snapshot, image, and volume", err)
	}

	instance := &types.Instance{
		Id: instanceId,
		Name: name,
		State: types.InstanceState_Pending,
		Infrastructure: types.Infrastructure_AWS,
		ImageId: image.Id,
		Created: time.Now(),
	}

	p.state.ModifyInstances(func(instances map[string]*types.Instance) error{
		instances[instance.Id] = instance
		return nil
	})

	logrus.WithFields(logrus.Fields{"instance": instance}).Infof("instance created succesfully")

	return instance, nil
}

func verifyMntsInput(image *types.Image, mntPointsToVolumeIds map[string]string) error {
	for _, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint == "/" {
			//ignore boot mount point
			continue
		}
		_, ok := mntPointsToVolumeIds[deviceMapping.MountPoint]
		if !ok {
			logrus.WithFields(logrus.Fields{"required-device-mappings": image.DeviceMappings}).Errorf("requied mount point missing: %s", deviceMapping.MountPoint)
			return lxerrors.New("required mount point missing from input", nil)
		}
	}
	return nil
}