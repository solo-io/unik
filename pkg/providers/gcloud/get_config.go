package gcloud

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *GcloudProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
