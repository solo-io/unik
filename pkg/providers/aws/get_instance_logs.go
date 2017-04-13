package aws

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/cf-unik/unik/pkg/providers/common"
)

func (p *AwsProvider) GetInstanceLogs(id string) (string, error) {
	instance, err := p.GetInstance(id)
	if err != nil {
		return "", errors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(instance)
}
