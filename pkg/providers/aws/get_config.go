package aws

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *AwsProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
		SupportedCompilers: []string{
			compilers.RUMP_GO_AWS,
			compilers.RUMP_NODEJS_AWS,
			compilers.RUMP_PYTHON_AWS,
			compilers.OSV_JAVA_AWS,
		},
	}
}
