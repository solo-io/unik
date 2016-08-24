package xen

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *XenProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
		SupportedCompilers: []string{
			compilers.RUMP_GO_XEN,
			compilers.RUMP_NODEJS_XEN,
			compilers.RUMP_PYTHON_XEN,
			compilers.RUMP_JAVA_XEN,
			compilers.OSV_JAVA_XEN,
		},
	}
}
