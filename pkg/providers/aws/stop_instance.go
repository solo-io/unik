package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *AwsProvider) StopInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	param := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	ec2svc := p.newEC2()
	_, err = ec2svc.StopInstances(param)
	if err != nil {
		return errors.New("failed to stop instance "+instance.Id, err)
	}
	waitParam := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(instance.Id)},
	}
	err = ec2svc.WaitUntilInstanceStopped(waitParam)
	if err != nil {
		return errors.New("waiting until instance stopped", err)
	}
	return nil
}
