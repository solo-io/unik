package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
)

func (p *AwsProvider) StartInstance(logger lxlog.Logger, id string) error {
	instance, err := p.GetInstance(logger, id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	param := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(instance.Id),
		},
	}
	_, err = p.newEC2(logger).StartInstances(param)
	if err != nil {
		return lxerrors.New("failed to start instance "+instance.Id, err)
	}
	return nil
}
