package gcloud

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *GcloudProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
