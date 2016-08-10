package aws

import (
	"encoding/base64"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"time"
)

func (p *AwsProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	var instanceId string
	ec2svc := p.newEC2()

	defer func() {
		if err != nil {
			logrus.WithError(err).Errorf("aws running instance encountered an error")
			if instanceId != "" {
				if params.NoCleanup {
					logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed instance %s0", instanceId)
					return
				}
				logrus.Warnf("cleaning up instance %s", instanceId)
				terminateInstanceInput := &ec2.TerminateInstancesInput{
					InstanceIds: []*string{aws.String(instanceId)},
				}
				ec2svc.TerminateInstances(terminateInstanceInput)
				cleanupErr := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
					delete(instances, instanceId)
					return nil
				})
				if cleanupErr != nil {
					logrus.Error(errors.New("modifying instance map in state", cleanupErr))
				}
				cleanupErr = p.state.Save()
				if cleanupErr != nil {
					logrus.Error(errors.New("saving instance volume map to state", cleanupErr))
				}
			}
		}
	}()

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	envData, err := json.Marshal(params.Env)
	if err != nil {
		return nil, errors.New("could not convert instance env to json", err)
	}
	encodedData := base64.StdEncoding.EncodeToString(envData)

	//if not set, use default
	if params.InstanceMemory <= 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	instanceType, err := getInstanceType(image.StageSpec.XenVirtualizationType, params.InstanceMemory)
	if err != nil {
		return nil, errors.New("could not find instance type for specified memory", err)
	}

	logrus.Debugf("determined intstance type %s for memory requirement %v", instanceType, params.InstanceMemory)

	runInstanceInput := &ec2.RunInstancesInput{
		ImageId:  aws.String(image.Id),
		MinCount: aws.Int64(1),
		MaxCount: aws.Int64(1),
		Placement: &ec2.Placement{
			AvailabilityZone: aws.String(p.config.Zone),
		},
		InstanceType: aws.String(instanceType),
		UserData:     aws.String(encodedData),
	}

	runInstanceOutput, err := ec2svc.RunInstances(runInstanceInput)
	if err != nil {
		return nil, errors.New("failed to run instance", err)
	}
	if len(runInstanceOutput.Instances) < 1 {
		logrus.WithFields(logrus.Fields{"output": runInstanceOutput}).Errorf("run instance %s failed, produced %v instances, expected 1", params.Name, len(runInstanceOutput.Instances))
		return nil, errors.New("expected 1 instance to be created", nil)
	}
	instanceId = *runInstanceOutput.Instances[0].InstanceId

	if len(runInstanceOutput.Instances) > 1 {
		logrus.WithFields(logrus.Fields{"output": runInstanceOutput}).Errorf("run instance %s failed, produced %v instances, expected 1", params.Name, len(runInstanceOutput.Instances))
		return nil, errors.New("expected 1 instance to be created", nil)
	}

	//must add instance to state before attaching volumes
	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		Infrastructure: types.Infrastructure_AWS,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("modifying instance map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return nil, errors.New("saving instance volume map to state", err)
	}

	if len(params.MntPointsToVolumeIds) > 0 {
		logrus.Debugf("stopping instance for volume attach")
		waitParam := &ec2.DescribeInstancesInput{
			InstanceIds: []*string{aws.String(instanceId)},
		}
		logrus.Debugf("waiting for instance to reach running state")
		if err := ec2svc.WaitUntilInstanceRunning(waitParam); err != nil {
			return nil, errors.New("waiting for instance to reach running state", err)
		}
		if err := p.StopInstance(instanceId); err != nil {
			return nil, errors.New("failed to stop instance for attaching volumes", err)
		}
		for mountPoint, volumeId := range params.MntPointsToVolumeIds {
			logrus.WithFields(logrus.Fields{"volume-id": volumeId}).Debugf("attaching volume %s to intance %s", volumeId, instanceId)
			if err := p.AttachVolume(volumeId, instanceId, mountPoint); err != nil {
				return nil, errors.New("attaching volume to instance", err)
			}
		}
		if err := p.StartInstance(instanceId); err != nil {
			return nil, errors.New("starting instance after volume attach", err)
		}
	}

	tagObjects := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(instanceId),
		},
		Tags: []*ec2.Tag{
			&ec2.Tag{
				Key:   aws.String("Name"),
				Value: aws.String(params.Name),
			},
		},
	}
	_, err = ec2svc.CreateTags(tagObjects)
	if err != nil {
		return nil, errors.New("tagging snapshot, image, and volume", err)
	}

	logrus.WithFields(logrus.Fields{"instance": instance}).Infof("instance created succesfully")

	return instance, nil
}

type instanceType struct {
	memory int
	name   string
}

var hvmInstanceTypes = []instanceType{
	instanceType{memory: 512, name: "t2.nano"},
	instanceType{memory: 1024, name: "t2.micro"},
	instanceType{memory: 2048, name: "t2.small"},
	instanceType{memory: 4096, name: "t2.medium"},
	instanceType{memory: 8192, name: "t2.large"},
	instanceType{memory: 16384, name: "m4.xlarge"},
}

var pvInstanceTypes = []instanceType{
	instanceType{memory: 1741, name: "m1.small"},
	instanceType{memory: 3789, name: "m1.medium"},
	instanceType{memory: 7680, name: "m1.large"},
	instanceType{memory: 15360, name: "m1.xlarge"},
}

func getInstanceType(virtualizationType types.XenVirtualizationType, memoryRequirement int) (string, error) {
	switch virtualizationType {
	case types.XenVirtualizationType_HVM:
		for _, instanceType := range hvmInstanceTypes {
			if instanceType.memory >= memoryRequirement {
				return instanceType.name, nil
			}
		}
		return "", errors.New("memory requirement too large", nil)
	case types.XenVirtualizationType_Paravirtual:
		for _, instanceType := range pvInstanceTypes {
			if instanceType.memory >= memoryRequirement {
				return instanceType.name, nil
			}
		}
		return "", errors.New("memory requirement too large", nil)
	}
	return "", errors.New("unknown virtualization type "+string(virtualizationType), nil)
}
