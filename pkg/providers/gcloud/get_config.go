package gcloud

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *GcloudProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
