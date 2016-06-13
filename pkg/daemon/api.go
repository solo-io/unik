package daemon

type RunInstanceRequest struct {
	InstanceName string `json:"InstanceName"`
	ImageName string `json:"ImageName"`
	Mounts map[string]string `json:"Mounts"`
	Env map[string]string `json:"Env"`
	MemoryMb int `json:"MemoryMb"`
	NoCleanup bool `json:"NoCleanup"`
	DebugMode bool `json:"DebugMode"`
}