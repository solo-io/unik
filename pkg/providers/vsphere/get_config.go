package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *VsphereProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
		SupportedCompilers: []string{
			compilers.RUMP_GO_VMWARE,
			compilers.RUMP_NODEJS_VMWARE,
			compilers.RUMP_PYTHON_VMWARE,
			compilers.OSV_JAVA_VMAWRE,
		},
	}
}
