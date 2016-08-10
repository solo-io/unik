package rump

type blk struct {
	Source     string `json:"source"`
	Path       string `json:"path"`
	FSType     string `json:"fstype"`
	MountPoint string `json:"mountpoint,omitempty"`
	DiskFile   string `json:"diskfile,omitempty"`
}

type Method string

const (
	Static Method = "static"
	DHCP   Method = "dhcp"
)

type net struct {
	If     string `json:"if,omitempty"`
	Type   string `json:"type,omitempty"`
	Method Method `json:"method,omitempty"`
	Addr   string `json:"addr,omitempty"`
	Mask   string `json:"mask,omitempty"`
	Cloner string `json:"cloner,omitempty"`
}

type commandLine struct {
	Bin     string   `json:"bin"`
	Argv    []string `json:"argv"`
	Runmode *string  `json:"runmode,omitempty"`
}

type rumpConfig struct {
	Rc   []commandLine     `json:"rc"`
	Net  *net              `json:"net,omitempty"`
	Net1 *net              `json:"net1,omitempty"`
	Blk  []blk             `json:"blk,omitempty"`
	Env  map[string]string `json:"env,omitempty"`
}
