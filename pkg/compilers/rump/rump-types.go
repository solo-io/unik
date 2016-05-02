package compilers

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

type rumpConfig struct {
	Cmdline string `json:"cmdline"`
	Net     *net   `json:"net,omitempty"`
	Blk     []blk  `json:"blk,omitempty"`
}

type multinetRumpConfig struct {
	Cmdline string `json:"cmdline"`
	Net1    *net   `json:"net1,omitempty"`
	Net2    *net   `json:"net2,omitempty"`
	Blk     []blk  `json:"blk,omitempty"`
}
