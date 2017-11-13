package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/providers/common"
)

func (p *GcloudProvider) GetInstanceLogs(id string) (string, error) {
	instance, err := p.GetInstance(id)
	if err != nil {
		return "", errors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(instance)
}
