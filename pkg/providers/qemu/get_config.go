package qemu

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *QemuProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
