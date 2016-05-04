package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func getStorageType(extraConfig types.ExtraConfig) string {
	var storageType string
	switch extraConfig[STORAGE_CONTROLLER_TYPE]{
	case SATA_Storage:
		storageType = SATA_Storage
	default:
		storageType = SCSI_Storage
	}
	return storageType
}