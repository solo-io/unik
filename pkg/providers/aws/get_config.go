package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func (p *AwsProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
		SupportedCompilers: []string{
			compilers.RUMP_GO_AWS,
			compilers.RUMP_NODEJS_AWS,
			compilers.OSV_JAVA_AWS,
		},
	}
}
