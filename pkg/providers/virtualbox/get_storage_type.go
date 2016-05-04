package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func getStorageType(extraConfig types.ExtraConfig) string {
	var storageType string
	switch extraConfig[STORAGE_CONTROLLER_TYPE]{
	case SATA_CONTROLLER:
		storageType = virtualboxclient.SATA_Storage
	default:
		storageType = virtualboxclient.SCSI_Storage
	}
	return storageType
}