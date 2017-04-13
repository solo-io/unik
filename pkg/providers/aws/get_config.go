package aws

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *AwsProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
