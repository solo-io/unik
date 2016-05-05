package aws

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"os"
)

var AwsStateFile = os.Getenv("HOME") + "/.unik/aws/state.json"

type AwsProvider struct {
	config config.Aws
	state  state.State
}

func NewAwsProvier(config config.Aws) *AwsProvider {
	logrus.Infof("state file: %s", AwsStateFile)
	return &AwsProvider{
		config: config,
		state:  state.NewBasicState(AwsStateFile),
	}
}

func (p *AwsProvider) WithState(state state.State) *AwsProvider {
	p.state = state
	return p
}

func (p *AwsProvider) newEC2() *ec2.EC2 {
	sess := session.New(&aws.Config{
		Region:      aws.String(p.config.Region),
	})
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		if r != nil {
			logrus.WithFields(logrus.Fields{"params": r.Params}).Debugf("request sent to ec2")
		}
	})
	return ec2.New(sess)
}

func (p *AwsProvider) newS3() *s3.S3 {
	sess := session.New(&aws.Config{
		Region:      aws.String(p.config.Region),
	})
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		if r != nil {
			logrus.WithFields(logrus.Fields{"params": r.Params}).Debugf("request sent to s3")
		}
	})
	return s3.New(sess)
}
