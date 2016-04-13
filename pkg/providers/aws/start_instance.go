package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *AwsProvider) StartInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	param := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	ec2svc := p.newEC2()
	_, err = ec2svc.StartInstances(param)
	if err != nil {
		return lxerrors.New("failed to start instance "+instance.Id, err)
	}
	waitParam := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instance.Id)},
	}
	err = ec2svc.WaitUntilInstanceRunning(waitParam)
	if err != nil {
		return lxerrors.New("waiting until instance running", err)
	}
	return nil
}
