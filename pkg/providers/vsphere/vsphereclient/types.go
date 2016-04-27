package vsphereclient

type VmInfo struct {
	VirtualMachines []VirtualMachine `json:"VirtualMachines"`
}

type VirtualMachine struct {
	Self struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"Self"`
	Value          interface{} `json:"Value"`
	AvailableField interface{} `json:"AvailableField"`
	Parent         struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"Parent"`
	CustomValue         interface{} `json:"CustomValue"`
	OverallStatus       string      `json:"OverallStatus"`
	ConfigStatus        string      `json:"ConfigStatus"`
	ConfigIssue         interface{} `json:"ConfigIssue"`
	EffectiveRole       []int       `json:"EffectiveRole"`
	Permission          interface{} `json:"Permission"`
	Name                string      `json:"Name"`
	DisabledMethod      []string    `json:"DisabledMethod"`
	RecentTask          interface{} `json:"RecentTask"`
	DeclaredAlarmState  interface{} `json:"DeclaredAlarmState"`
	TriggeredAlarmState interface{} `json:"TriggeredAlarmState"`
	AlarmActionsEnabled interface{} `json:"AlarmActionsEnabled"`
	Tag                 interface{} `json:"Tag"`
	Capability          struct {
		SnapshotOperationsSupported         bool `json:"SnapshotOperationsSupported"`
		MultipleSnapshotsSupported          bool `json:"MultipleSnapshotsSupported"`
		SnapshotConfigSupported             bool `json:"SnapshotConfigSupported"`
		PoweredOffSnapshotsSupported        bool `json:"PoweredOffSnapshotsSupported"`
		MemorySnapshotsSupported            bool `json:"MemorySnapshotsSupported"`
		RevertToSnapshotSupported           bool `json:"RevertToSnapshotSupported"`
		QuiescedSnapshotsSupported          bool `json:"QuiescedSnapshotsSupported"`
		DisableSnapshotsSupported           bool `json:"DisableSnapshotsSupported"`
		LockSnapshotsSupported              bool `json:"LockSnapshotsSupported"`
		ConsolePreferencesSupported         bool `json:"ConsolePreferencesSupported"`
		CPUFeatureMaskSupported             bool `json:"CpuFeatureMaskSupported"`
		S1AcpiManagementSupported           bool `json:"S1AcpiManagementSupported"`
		SettingScreenResolutionSupported    bool `json:"SettingScreenResolutionSupported"`
		ToolsAutoUpdateSupported            bool `json:"ToolsAutoUpdateSupported"`
		VMNpivWwnSupported                  bool `json:"VmNpivWwnSupported"`
		NpivWwnOnNonRdmVMSupported          bool `json:"NpivWwnOnNonRdmVmSupported"`
		VMNpivWwnDisableSupported           bool `json:"VmNpivWwnDisableSupported"`
		VMNpivWwnUpdateSupported            bool `json:"VmNpivWwnUpdateSupported"`
		SwapPlacementSupported              bool `json:"SwapPlacementSupported"`
		ToolsSyncTimeSupported              bool `json:"ToolsSyncTimeSupported"`
		VirtualMmuUsageSupported            bool `json:"VirtualMmuUsageSupported"`
		DiskSharesSupported                 bool `json:"DiskSharesSupported"`
		BootOptionsSupported                bool `json:"BootOptionsSupported"`
		BootRetryOptionsSupported           bool `json:"BootRetryOptionsSupported"`
		SettingVideoRAMSizeSupported        bool `json:"SettingVideoRamSizeSupported"`
		SettingDisplayTopologySupported     bool `json:"SettingDisplayTopologySupported"`
		RecordReplaySupported               bool `json:"RecordReplaySupported"`
		ChangeTrackingSupported             bool `json:"ChangeTrackingSupported"`
		MultipleCoresPerSocketSupported     bool `json:"MultipleCoresPerSocketSupported"`
		HostBasedReplicationSupported       bool `json:"HostBasedReplicationSupported"`
		GuestAutoLockSupported              bool `json:"GuestAutoLockSupported"`
		MemoryReservationLockSupported      bool `json:"MemoryReservationLockSupported"`
		FeatureRequirementSupported         bool `json:"FeatureRequirementSupported"`
		PoweredOnMonitorTypeChangeSupported bool `json:"PoweredOnMonitorTypeChangeSupported"`
		SeSparseDiskSupported               bool `json:"SeSparseDiskSupported"`
		NestedHVSupported                   bool `json:"NestedHVSupported"`
		VPMCSupported                       bool `json:"VPMCSupported"`
	} `json:"Capability"`
	Config struct {
		Name                  string      `json:"Name"`
		GuestFullName         string      `json:"GuestFullName"`
		Version               string      `json:"Version"`
		UUID                  string      `json:"Uuid"`
		InstanceUUID          string      `json:"InstanceUuid"`
		NpivNodeWorldWideName interface{} `json:"NpivNodeWorldWideName"`
		NpivPortWorldWideName interface{} `json:"NpivPortWorldWideName"`
		NpivWorldWideNameType string      `json:"NpivWorldWideNameType"`
		NpivDesiredNodeWwns   int         `json:"NpivDesiredNodeWwns"`
		NpivDesiredPortWwns   int         `json:"NpivDesiredPortWwns"`
		NpivTemporaryDisabled bool        `json:"NpivTemporaryDisabled"`
		NpivOnNonRdmDisks     interface{} `json:"NpivOnNonRdmDisks"`
		LocationID            string      `json:"LocationId"`
		Template              bool        `json:"Template"`
		GuestID               string      `json:"GuestId"`
		AlternateGuestName    string      `json:"AlternateGuestName"`
		Annotation            string      `json:"Annotation"`
		Files                 struct {
			VMPathName          string `json:"VmPathName"`
			SnapshotDirectory   string `json:"SnapshotDirectory"`
			SuspendDirectory    string `json:"SuspendDirectory"`
			LogDirectory        string `json:"LogDirectory"`
			FtMetadataDirectory string `json:"FtMetadataDirectory"`
		} `json:"Files"`
		Tools struct {
			ToolsVersion         int         `json:"ToolsVersion"`
			AfterPowerOn         bool        `json:"AfterPowerOn"`
			AfterResume          bool        `json:"AfterResume"`
			BeforeGuestStandby   bool        `json:"BeforeGuestStandby"`
			BeforeGuestShutdown  bool        `json:"BeforeGuestShutdown"`
			BeforeGuestReboot    interface{} `json:"BeforeGuestReboot"`
			ToolsUpgradePolicy   string      `json:"ToolsUpgradePolicy"`
			PendingCustomization string      `json:"PendingCustomization"`
			SyncTimeWithHost     bool        `json:"SyncTimeWithHost"`
			LastInstallInfo      struct {
				Counter int         `json:"Counter"`
				Fault   interface{} `json:"Fault"`
			} `json:"LastInstallInfo"`
		} `json:"Tools"`
		Flags struct {
			DisableAcceleration      bool   `json:"DisableAcceleration"`
			EnableLogging            bool   `json:"EnableLogging"`
			UseToe                   bool   `json:"UseToe"`
			RunWithDebugInfo         bool   `json:"RunWithDebugInfo"`
			MonitorType              string `json:"MonitorType"`
			HtSharing                string `json:"HtSharing"`
			SnapshotDisabled         bool   `json:"SnapshotDisabled"`
			SnapshotLocked           bool   `json:"SnapshotLocked"`
			DiskUUIDEnabled          bool   `json:"DiskUuidEnabled"`
			VirtualMmuUsage          string `json:"VirtualMmuUsage"`
			VirtualExecUsage         string `json:"VirtualExecUsage"`
			SnapshotPowerOffBehavior string `json:"SnapshotPowerOffBehavior"`
			RecordReplayEnabled      bool   `json:"RecordReplayEnabled"`
			FaultToleranceType       string `json:"FaultToleranceType"`
		} `json:"Flags"`
		ConsolePreferences interface{} `json:"ConsolePreferences"`
		DefaultPowerOps    struct {
			PowerOffType        string `json:"PowerOffType"`
			SuspendType         string `json:"SuspendType"`
			ResetType           string `json:"ResetType"`
			DefaultPowerOffType string `json:"DefaultPowerOffType"`
			DefaultSuspendType  string `json:"DefaultSuspendType"`
			DefaultResetType    string `json:"DefaultResetType"`
			StandbyAction       string `json:"StandbyAction"`
		} `json:"DefaultPowerOps"`
		Hardware struct {
			NumCPU              int  `json:"NumCPU"`
			NumCoresPerSocket   int  `json:"NumCoresPerSocket"`
			MemoryMB            int  `json:"MemoryMB"`
			VirtualICH7MPresent bool `json:"VirtualICH7MPresent"`
			VirtualSMCPresent   bool `json:"VirtualSMCPresent"`
			Device              []struct {
				Key        int `json:"Key"`
				DeviceInfo struct {
					Label   string `json:"Label"`
					Summary string `json:"Summary"`
				} `json:"DeviceInfo"`
				Backing                        interface{} `json:"Backing"`
				Connectable                    interface{} `json:"Connectable"`
				SlotInfo                       interface{} `json:"SlotInfo"`
				ControllerKey                  int         `json:"ControllerKey"`
				UnitNumber                     interface{} `json:"UnitNumber"`
				BusNumber                      int         `json:"BusNumber,omitempty"`
				Device                         interface{} `json:"Device,omitempty"`
				VideoRAMSizeInKB               int         `json:"VideoRamSizeInKB,omitempty"`
				NumDisplays                    int         `json:"NumDisplays,omitempty"`
				UseAutoDetect                  bool        `json:"UseAutoDetect,omitempty"`
				Enable3DSupport                bool        `json:"Enable3DSupport,omitempty"`
				Use3DRenderer                  string      `json:"Use3dRenderer,omitempty"`
				GraphicsMemorySizeInKB         int         `json:"GraphicsMemorySizeInKB,omitempty"`
				ID                             int         `json:"Id,omitempty"`
				AllowUnrestrictedCommunication bool        `json:"AllowUnrestrictedCommunication,omitempty"`
				FilterEnable                   bool        `json:"FilterEnable,omitempty"`
				FilterInfo                     interface{} `json:"FilterInfo,omitempty"`
				HotAddRemove                   bool        `json:"HotAddRemove,omitempty"`
				SharedBus                      string      `json:"SharedBus,omitempty"`
				ScsiCtlrUnitNumber             int         `json:"ScsiCtlrUnitNumber,omitempty"`
				CapacityInKB                   int         `json:"CapacityInKB,omitempty"`
				CapacityInBytes                int64       `json:"CapacityInBytes,omitempty"`
				Shares                         struct {
					Shares int    `json:"Shares"`
					Level  string `json:"Level"`
				} `json:"Shares,omitempty"`
				StorageIOAllocation struct {
					Limit  int `json:"Limit"`
					Shares struct {
						Shares int    `json:"Shares"`
						Level  string `json:"Level"`
					} `json:"Shares"`
					Reservation int `json:"Reservation"`
				} `json:"StorageIOAllocation,omitempty"`
				DiskObjectID          string      `json:"DiskObjectId,omitempty"`
				VFlashCacheConfigInfo interface{} `json:"VFlashCacheConfigInfo,omitempty"`
				Iofilter              interface{} `json:"Iofilter,omitempty"`
				AddressType           string      `json:"AddressType,omitempty"`
				MacAddress            string      `json:"MacAddress,omitempty"`
				WakeOnLanEnabled      bool        `json:"WakeOnLanEnabled,omitempty"`
				ResourceAllocation    struct {
					Reservation int `json:"Reservation"`
					Share       struct {
						Shares int    `json:"Shares"`
						Level  string `json:"Level"`
					} `json:"Share"`
					Limit int `json:"Limit"`
				} `json:"ResourceAllocation,omitempty"`
				ExternalID              string `json:"ExternalId,omitempty"`
				UptCompatibilityEnabled bool   `json:"UptCompatibilityEnabled,omitempty"`
			} `json:"Device"`
		} `json:"Hardware"`
		CPUAllocation struct {
			Reservation           int  `json:"Reservation"`
			ExpandableReservation bool `json:"ExpandableReservation"`
			Limit                 int  `json:"Limit"`
			Shares                struct {
				Shares int    `json:"Shares"`
				Level  string `json:"Level"`
			} `json:"Shares"`
			OverheadLimit int `json:"OverheadLimit"`
		} `json:"CpuAllocation"`
		MemoryAllocation struct {
			Reservation           int  `json:"Reservation"`
			ExpandableReservation bool `json:"ExpandableReservation"`
			Limit                 int  `json:"Limit"`
			Shares                struct {
				Shares int    `json:"Shares"`
				Level  string `json:"Level"`
			} `json:"Shares"`
			OverheadLimit int `json:"OverheadLimit"`
		} `json:"MemoryAllocation"`
		LatencySensitivity struct {
			Level       string `json:"Level"`
			Sensitivity int    `json:"Sensitivity"`
		} `json:"LatencySensitivity"`
		MemoryHotAddEnabled        bool        `json:"MemoryHotAddEnabled"`
		CPUHotAddEnabled           bool        `json:"CpuHotAddEnabled"`
		CPUHotRemoveEnabled        bool        `json:"CpuHotRemoveEnabled"`
		HotPlugMemoryLimit         int         `json:"HotPlugMemoryLimit"`
		HotPlugMemoryIncrementSize int         `json:"HotPlugMemoryIncrementSize"`
		CPUAffinity                interface{} `json:"CpuAffinity"`
		MemoryAffinity             interface{} `json:"MemoryAffinity"`
		NetworkShaper              interface{} `json:"NetworkShaper"`
		ExtraConfig                []struct {
			Key   string `json:"Key"`
			Value string `json:"Value"`
		} `json:"ExtraConfig"`
		CPUFeatureMask []struct {
			Level  int    `json:"Level"`
			Vendor string `json:"Vendor"`
			Eax    string `json:"Eax"`
			Ebx    string `json:"Ebx"`
			Ecx    string `json:"Ecx"`
			Edx    string `json:"Edx"`
		} `json:"CpuFeatureMask"`
		DatastoreURL []struct {
			Name string `json:"Name"`
			URL  string `json:"Url"`
		} `json:"DatastoreUrl"`
		SwapPlacement string `json:"SwapPlacement"`
		BootOptions   struct {
			BootDelay           int         `json:"BootDelay"`
			EnterBIOSSetup      bool        `json:"EnterBIOSSetup"`
			BootRetryEnabled    bool        `json:"BootRetryEnabled"`
			BootRetryDelay      int         `json:"BootRetryDelay"`
			BootOrder           interface{} `json:"BootOrder"`
			NetworkBootProtocol string      `json:"NetworkBootProtocol"`
		} `json:"BootOptions"`
		FtInfo                       interface{} `json:"FtInfo"`
		RepConfig                    interface{} `json:"RepConfig"`
		VAppConfig                   interface{} `json:"VAppConfig"`
		VAssertsEnabled              bool        `json:"VAssertsEnabled"`
		ChangeTrackingEnabled        bool        `json:"ChangeTrackingEnabled"`
		Firmware                     string      `json:"Firmware"`
		MaxMksConnections            int         `json:"MaxMksConnections"`
		GuestAutoLockEnabled         bool        `json:"GuestAutoLockEnabled"`
		ManagedBy                    interface{} `json:"ManagedBy"`
		MemoryReservationLockedToMax bool        `json:"MemoryReservationLockedToMax"`
		InitialOverhead              struct {
			InitialMemoryReservation int `json:"InitialMemoryReservation"`
			InitialSwapReservation   int `json:"InitialSwapReservation"`
		} `json:"InitialOverhead"`
		NestedHVEnabled              bool `json:"NestedHVEnabled"`
		VPMCEnabled                  bool `json:"VPMCEnabled"`
		ScheduledHardwareUpgradeInfo struct {
			UpgradePolicy                  string      `json:"UpgradePolicy"`
			VersionKey                     string      `json:"VersionKey"`
			ScheduledHardwareUpgradeStatus string      `json:"ScheduledHardwareUpgradeStatus"`
			Fault                          interface{} `json:"Fault"`
		} `json:"ScheduledHardwareUpgradeInfo"`
		ForkConfigInfo struct {
			ParentEnabled    bool   `json:"ParentEnabled"`
			ChildForkGroupID string `json:"ChildForkGroupId"`
			ChildType        string `json:"ChildType"`
		} `json:"ForkConfigInfo"`
		VFlashCacheReservation  int    `json:"VFlashCacheReservation"`
		VmxConfigChecksum       string `json:"VmxConfigChecksum"`
		MessageBusTunnelEnabled bool   `json:"MessageBusTunnelEnabled"`
		VMStorageObjectID       string `json:"VmStorageObjectId"`
		SwapStorageObjectID     string `json:"SwapStorageObjectId"`
	} `json:"Config"`
	Layout struct {
		ConfigFile []string `json:"ConfigFile"`
		LogFile    []string `json:"LogFile"`
		Disk       []struct {
			Key      int      `json:"Key"`
			DiskFile []string `json:"DiskFile"`
		} `json:"Disk"`
		Snapshot interface{} `json:"Snapshot"`
		SwapFile string      `json:"SwapFile"`
	} `json:"Layout"`
	LayoutEx struct {
		File []struct {
			Key             int    `json:"Key"`
			Name            string `json:"Name"`
			Type            string `json:"Type"`
			Size            int    `json:"Size"`
			UniqueSize      int    `json:"UniqueSize"`
			BackingObjectID string `json:"BackingObjectId"`
			Accessible      bool   `json:"Accessible"`
		} `json:"File"`
		Disk []struct {
			Key   int `json:"Key"`
			Chain []struct {
				FileKey []int `json:"FileKey"`
			} `json:"Chain"`
		} `json:"Disk"`
		Snapshot interface{} `json:"Snapshot"`
	} `json:"LayoutEx"`
	Storage struct {
		PerDatastoreUsage []struct {
			Datastore struct {
				Type  string `json:"Type"`
				Value string `json:"Value"`
			} `json:"Datastore"`
			Committed   int   `json:"Committed"`
			Uncommitted int64 `json:"Uncommitted"`
			Unshared    int   `json:"Unshared"`
		} `json:"PerDatastoreUsage"`
	} `json:"Storage"`
	EnvironmentBrowser struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"EnvironmentBrowser"`
	ResourcePool struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"ResourcePool"`
	ParentVApp     interface{} `json:"ParentVApp"`
	ResourceConfig struct {
		Entity struct {
			Type  string `json:"Type"`
			Value string `json:"Value"`
		} `json:"Entity"`
		ChangeVersion string      `json:"ChangeVersion"`
		LastModified  interface{} `json:"LastModified"`
		CPUAllocation struct {
			Reservation           int  `json:"Reservation"`
			ExpandableReservation bool `json:"ExpandableReservation"`
			Limit                 int  `json:"Limit"`
			Shares                struct {
				Shares int    `json:"Shares"`
				Level  string `json:"Level"`
			} `json:"Shares"`
			OverheadLimit int `json:"OverheadLimit"`
		} `json:"CpuAllocation"`
		MemoryAllocation struct {
			Reservation           int  `json:"Reservation"`
			ExpandableReservation bool `json:"ExpandableReservation"`
			Limit                 int  `json:"Limit"`
			Shares                struct {
				Shares int    `json:"Shares"`
				Level  string `json:"Level"`
			} `json:"Shares"`
			OverheadLimit int `json:"OverheadLimit"`
		} `json:"MemoryAllocation"`
	} `json:"ResourceConfig"`
	Runtime struct {
		Device []struct {
			RuntimeState struct {
				VMDirectPathGen2Active                 bool     `json:"VmDirectPathGen2Active"`
				VMDirectPathGen2InactiveReasonVM       []string `json:"VmDirectPathGen2InactiveReasonVm"`
				VMDirectPathGen2InactiveReasonOther    []string `json:"VmDirectPathGen2InactiveReasonOther"`
				VMDirectPathGen2InactiveReasonExtended string   `json:"VmDirectPathGen2InactiveReasonExtended"`
				ReservationStatus                      string   `json:"ReservationStatus"`
			} `json:"RuntimeState"`
			Key int `json:"Key"`
		} `json:"Device"`
		Host struct {
			Type  string `json:"Type"`
			Value string `json:"Value"`
		} `json:"Host"`
		ConnectionState           string      `json:"ConnectionState"`
		PowerState                string      `json:"PowerState"`
		FaultToleranceState       string      `json:"FaultToleranceState"`
		DasVMProtection           interface{} `json:"DasVmProtection"`
		ToolsInstallerMounted     bool        `json:"ToolsInstallerMounted"`
		SuspendTime               interface{} `json:"SuspendTime"`
		BootTime                  interface{} `json:"BootTime"`
		SuspendInterval           int         `json:"SuspendInterval"`
		Question                  interface{} `json:"Question"`
		MemoryOverhead            int         `json:"MemoryOverhead"`
		MaxCPUUsage               int         `json:"MaxCpuUsage"`
		MaxMemoryUsage            int         `json:"MaxMemoryUsage"`
		NumMksConnections         int         `json:"NumMksConnections"`
		RecordReplayState         string      `json:"RecordReplayState"`
		CleanPowerOff             bool        `json:"CleanPowerOff"`
		NeedSecondaryReason       string      `json:"NeedSecondaryReason"`
		OnlineStandby             bool        `json:"OnlineStandby"`
		MinRequiredEVCModeKey     string      `json:"MinRequiredEVCModeKey"`
		ConsolidationNeeded       bool        `json:"ConsolidationNeeded"`
		OfflineFeatureRequirement interface{} `json:"OfflineFeatureRequirement"`
		FeatureRequirement        interface{} `json:"FeatureRequirement"`
		FeatureMask               interface{} `json:"FeatureMask"`
		VFlashCacheAllocation     int         `json:"VFlashCacheAllocation"`
		Paused                    bool        `json:"Paused"`
		SnapshotInBackground      bool        `json:"SnapshotInBackground"`
		QuiescedForkParent        interface{} `json:"QuiescedForkParent"`
	} `json:"Runtime"`
	Guest struct {
		ToolsStatus         string      `json:"ToolsStatus"`
		ToolsVersionStatus  string      `json:"ToolsVersionStatus"`
		ToolsVersionStatus2 string      `json:"ToolsVersionStatus2"`
		ToolsRunningStatus  string      `json:"ToolsRunningStatus"`
		ToolsVersion        string      `json:"ToolsVersion"`
		GuestID             string      `json:"GuestId"`
		GuestFamily         string      `json:"GuestFamily"`
		GuestFullName       string      `json:"GuestFullName"`
		HostName            string      `json:"HostName"`
		IPAddress           string      `json:"IpAddress"`
		Net                 interface{} `json:"Net"`
		IPStack             interface{} `json:"IpStack"`
		Disk                interface{} `json:"Disk"`
		Screen              struct {
			Width  int `json:"Width"`
			Height int `json:"Height"`
		} `json:"Screen"`
		GuestState                      string      `json:"GuestState"`
		AppHeartbeatStatus              string      `json:"AppHeartbeatStatus"`
		GuestKernelCrashed              interface{} `json:"GuestKernelCrashed"`
		AppState                        string      `json:"AppState"`
		GuestOperationsReady            bool        `json:"GuestOperationsReady"`
		InteractiveGuestOperationsReady bool        `json:"InteractiveGuestOperationsReady"`
		GuestStateChangeSupported       bool        `json:"GuestStateChangeSupported"`
		GenerationInfo                  interface{} `json:"GenerationInfo"`
	} `json:"Guest"`
	Summary struct {
		VM struct {
			Type  string `json:"Type"`
			Value string `json:"Value"`
		} `json:"Vm"`
		Runtime struct {
			Device []struct {
				RuntimeState struct {
					VMDirectPathGen2Active                 bool     `json:"VmDirectPathGen2Active"`
					VMDirectPathGen2InactiveReasonVM       []string `json:"VmDirectPathGen2InactiveReasonVm"`
					VMDirectPathGen2InactiveReasonOther    []string `json:"VmDirectPathGen2InactiveReasonOther"`
					VMDirectPathGen2InactiveReasonExtended string   `json:"VmDirectPathGen2InactiveReasonExtended"`
					ReservationStatus                      string   `json:"ReservationStatus"`
				} `json:"RuntimeState"`
				Key int `json:"Key"`
			} `json:"Device"`
			Host struct {
				Type  string `json:"Type"`
				Value string `json:"Value"`
			} `json:"Host"`
			ConnectionState           string      `json:"ConnectionState"`
			PowerState                string      `json:"PowerState"`
			FaultToleranceState       string      `json:"FaultToleranceState"`
			DasVMProtection           interface{} `json:"DasVmProtection"`
			ToolsInstallerMounted     bool        `json:"ToolsInstallerMounted"`
			SuspendTime               interface{} `json:"SuspendTime"`
			BootTime                  interface{} `json:"BootTime"`
			SuspendInterval           int         `json:"SuspendInterval"`
			Question                  interface{} `json:"Question"`
			MemoryOverhead            int         `json:"MemoryOverhead"`
			MaxCPUUsage               int         `json:"MaxCpuUsage"`
			MaxMemoryUsage            int         `json:"MaxMemoryUsage"`
			NumMksConnections         int         `json:"NumMksConnections"`
			RecordReplayState         string      `json:"RecordReplayState"`
			CleanPowerOff             bool        `json:"CleanPowerOff"`
			NeedSecondaryReason       string      `json:"NeedSecondaryReason"`
			OnlineStandby             bool        `json:"OnlineStandby"`
			MinRequiredEVCModeKey     string      `json:"MinRequiredEVCModeKey"`
			ConsolidationNeeded       bool        `json:"ConsolidationNeeded"`
			OfflineFeatureRequirement interface{} `json:"OfflineFeatureRequirement"`
			FeatureRequirement        interface{} `json:"FeatureRequirement"`
			FeatureMask               interface{} `json:"FeatureMask"`
			VFlashCacheAllocation     int         `json:"VFlashCacheAllocation"`
			Paused                    bool        `json:"Paused"`
			SnapshotInBackground      bool        `json:"SnapshotInBackground"`
			QuiescedForkParent        interface{} `json:"QuiescedForkParent"`
		} `json:"Runtime"`
		Guest struct {
			GuestID             string `json:"GuestId"`
			GuestFullName       string `json:"GuestFullName"`
			ToolsStatus         string `json:"ToolsStatus"`
			ToolsVersionStatus  string `json:"ToolsVersionStatus"`
			ToolsVersionStatus2 string `json:"ToolsVersionStatus2"`
			ToolsRunningStatus  string `json:"ToolsRunningStatus"`
			HostName            string `json:"HostName"`
			IPAddress           string `json:"IpAddress"`
		} `json:"Guest"`
		Config struct {
			Name                string      `json:"Name"`
			Template            bool        `json:"Template"`
			VMPathName          string      `json:"VmPathName"`
			MemorySizeMB        int         `json:"MemorySizeMB"`
			CPUReservation      int         `json:"CpuReservation"`
			MemoryReservation   int         `json:"MemoryReservation"`
			NumCPU              int         `json:"NumCpu"`
			NumEthernetCards    int         `json:"NumEthernetCards"`
			NumVirtualDisks     int         `json:"NumVirtualDisks"`
			UUID                string      `json:"Uuid"`
			InstanceUUID        string      `json:"InstanceUuid"`
			GuestID             string      `json:"GuestId"`
			GuestFullName       string      `json:"GuestFullName"`
			Annotation          string      `json:"Annotation"`
			Product             interface{} `json:"Product"`
			InstallBootRequired interface{} `json:"InstallBootRequired"`
			FtInfo              interface{} `json:"FtInfo"`
			ManagedBy           interface{} `json:"ManagedBy"`
		} `json:"Config"`
		Storage struct {
			Committed   int   `json:"Committed"`
			Uncommitted int64 `json:"Uncommitted"`
			Unshared    int   `json:"Unshared"`
		} `json:"Storage"`
		QuickStats struct {
			OverallCPUUsage              int    `json:"OverallCpuUsage"`
			OverallCPUDemand             int    `json:"OverallCpuDemand"`
			GuestMemoryUsage             int    `json:"GuestMemoryUsage"`
			HostMemoryUsage              int    `json:"HostMemoryUsage"`
			GuestHeartbeatStatus         string `json:"GuestHeartbeatStatus"`
			DistributedCPUEntitlement    int    `json:"DistributedCpuEntitlement"`
			DistributedMemoryEntitlement int    `json:"DistributedMemoryEntitlement"`
			StaticCPUEntitlement         int    `json:"StaticCpuEntitlement"`
			StaticMemoryEntitlement      int    `json:"StaticMemoryEntitlement"`
			PrivateMemory                int    `json:"PrivateMemory"`
			SharedMemory                 int    `json:"SharedMemory"`
			SwappedMemory                int    `json:"SwappedMemory"`
			BalloonedMemory              int    `json:"BalloonedMemory"`
			ConsumedOverheadMemory       int    `json:"ConsumedOverheadMemory"`
			FtLogBandwidth               int    `json:"FtLogBandwidth"`
			FtSecondaryLatency           int    `json:"FtSecondaryLatency"`
			FtLatencyStatus              string `json:"FtLatencyStatus"`
			CompressedMemory             int    `json:"CompressedMemory"`
			UptimeSeconds                int    `json:"UptimeSeconds"`
			SsdSwappedMemory             int    `json:"SsdSwappedMemory"`
		} `json:"QuickStats"`
		OverallStatus string      `json:"OverallStatus"`
		CustomValue   interface{} `json:"CustomValue"`
	} `json:"Summary"`
	Datastore []struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"Datastore"`
	Network []struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	} `json:"Network"`
	Snapshot             interface{} `json:"Snapshot"`
	RootSnapshot         interface{} `json:"RootSnapshot"`
	GuestHeartbeatStatus string      `json:"GuestHeartbeatStatus"`
}
