package aws

import "github.com/emc-advanced-dev/unik/pkg/providers"

func (p *AwsProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
