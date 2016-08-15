package xenclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
)

const xenConfBase = `
# Example PV Linux guest configuration
# =====================================================================
#
# This is a fairly minimal example of what is required for a
# Paravirtualised Linux guest. For a more complete guide see xl.cfg(5)

# Guest name
name = "%s"

kernel = "%s"
extra = "(hd0)/boot/grub/menu.lst"

# Initial memory allocation (MB)
memory = %d

# Number of VCPUS
vcpus = 1

# Network devices
# A list of 'vifspec' entries as described in
# docs/misc/xl-network-configuration.markdown
vif = [ 'bridge=%s' ]

# Disk Devices
# A list of 'diskspec' entries as described in
# docs/misc/xl-disk-configuration.txt
disk = [ '%s,raw,sda1,rw'%s ]

on_poweroff = "preserve"
on_reboot = "preserve"
on_crash = "preserve"
`

type XenClient struct {
	KernelPath string
	XenBridge  string
}

type CreateVmParams struct {
	Name        string
	Memory      int
	BootImage   string
	VmDir       string
	DataVolumes []VolumeConfig
}

type VolumeConfig struct {
	ImagePath  string
	DeviceName string
}

func (c *XenClient) CreateVm(params CreateVmParams) error {
	volumes := ""
	for _, vol := range params.DataVolumes {
		volumes = fmt.Sprintf("%s, '%s,raw,%s,rw'", volumes, vol.ImagePath, vol.DeviceName)
	}
	xenConf := fmt.Sprintf(xenConfBase, params.Name, c.KernelPath, params.Memory, c.XenBridge, params.BootImage, volumes)
	confFile := filepath.Join(params.VmDir, "xen.conf")
	if err := ioutil.WriteFile(confFile, []byte(xenConf), 0644); err != nil {
		return errors.New("writing xen conf file for vm", err)
	}

	logrus.Debugf("using xen config:\n%s", xenConf)

	if _, err := xl("create", confFile); err != nil {
		return errors.New("creating domain", err)
	}
	return nil
}

func (c *XenClient) DestroyVm(name string) error {
	if _, err := xl("destroy", name); err != nil {
		return errors.New("destroying domain", err)
	}
	return nil
}

func (c *XenClient) ListVms() (domList, error) {
	res, err := xl("list", "-l")
	if err != nil {
		return nil, errors.New("getting xl list", err)
	}
	var domList domList
	if err := json.Unmarshal(res, &domList); err != nil {
		return nil, errors.New("parsing "+string(res)+" as domain list", err)
	}
	return domList, nil
}

func xl(args ...string) ([]byte, error) {
	cmd := exec.Command("xl", args...)
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf
	out, err := cmd.Output()
	if err != nil {
		return out, errors.New(errBuf.String(), err)
	}
	return out, nil
}

type domList []struct {
	Config struct {
		OnCrash    string `json:"on_crash"`
		OnWatchdog string `json:"on_watchdog"`
		OnReboot   string `json:"on_reboot"`
		OnPoweroff string `json:"on_poweroff"`
		CInfo      struct {
			DriverDomain      string `json:"driver_domain"`
			Pvh               string `json:"pvh"`
			RunHotplugScripts string `json:"run_hotplug_scripts"`
			Poolid            int    `json:"poolid"`
			Type              string `json:"type"`
			Hap               string `json:"hap"`
			Oos               string `json:"oos"`
			Ssidref           int    `json:"ssidref"`
			Name              string `json:"name"`
			UUID              string `json:"uuid"`
			Xsdata            struct {
			} `json:"xsdata"`
			Platformdata struct {
			} `json:"platformdata"`
		} `json:"c_info"`
		BInfo struct {
			U struct {
				E820Host       string        `json:"e820_host"`
				Ramdisk        interface{}   `json:"ramdisk"`
				Cmdline        string        `json:"cmdline"`
				BootloaderArgs []interface{} `json:"bootloader_args"`
				Bootloader     interface{}   `json:"bootloader"`
				SlackMemkb     int           `json:"slack_memkb"`
				Kernel         string        `json:"kernel"`
			} `json:"u"`
			EventChannels int           `json:"event_channels"`
			ClaimMode     string        `json:"claim_mode"`
			Iomem         []interface{} `json:"iomem"`
			Irqs          []interface{} `json:"irqs"`
			Ioports       []interface{} `json:"ioports"`
			SchedParams   struct {
				Extratime int    `json:"extratime"`
				Latency   int    `json:"latency"`
				Slice     int    `json:"slice"`
				Period    int    `json:"period"`
				Cap       int    `json:"cap"`
				Weight    int    `json:"weight"`
				Sched     string `json:"sched"`
			} `json:"sched_params"`
			ExtraHvm              []interface{} `json:"extra_hvm"`
			ExtraPv               []interface{} `json:"extra_pv"`
			Extra                 []interface{} `json:"extra"`
			DeviceModelSsidref    int           `json:"device_model_ssidref"`
			DeviceModel           interface{}   `json:"device_model"`
			DeviceModelStubdomain string        `json:"device_model_stubdomain"`
			DeviceModelVersion    string        `json:"device_model_version"`
			TargetMemkb           int           `json:"target_memkb"`
			MaxMemkb              int           `json:"max_memkb"`
			TscMode               string        `json:"tsc_mode"`
			NumaPlacement         string        `json:"numa_placement"`
			Nodemap               []interface{} `json:"nodemap"`
			Cpumap                []interface{} `json:"cpumap"`
			AvailVcpus            []int         `json:"avail_vcpus"`
			MaxVcpus              int           `json:"max_vcpus"`
			VideoMemkb            int           `json:"video_memkb"`
			ShadowMemkb           int           `json:"shadow_memkb"`
			RtcTimeoffset         int           `json:"rtc_timeoffset"`
			ExecSsidref           int           `json:"exec_ssidref"`
			Localtime             string        `json:"localtime"`
			DisableMigrate        string        `json:"disable_migrate"`
			Cpuid                 []interface{} `json:"cpuid"`
			BlkdevStart           interface{}   `json:"blkdev_start"`
		} `json:"b_info"`
		Disks []struct {
			IsCdrom        int         `json:"is_cdrom"`
			Readwrite      int         `json:"readwrite"`
			BackendDomid   int         `json:"backend_domid"`
			BackendDomname interface{} `json:"backend_domname"`
			PdevPath       string      `json:"pdev_path"`
			Vdev           string      `json:"vdev"`
			Backend        string      `json:"backend"`
			Format         string      `json:"format"`
			Script         interface{} `json:"script"`
			Removable      int         `json:"removable"`
		} `json:"disks"`
		Nics []struct {
			Gatewaydev           interface{} `json:"gatewaydev"`
			RateIntervalUsecs    int         `json:"rate_interval_usecs"`
			RateBytesPerInterval int         `json:"rate_bytes_per_interval"`
			Nictype              string      `json:"nictype"`
			Script               interface{} `json:"script"`
			Ifname               interface{} `json:"ifname"`
			BackendDomid         int         `json:"backend_domid"`
			BackendDomname       interface{} `json:"backend_domname"`
			Devid                int         `json:"devid"`
			Mtu                  int         `json:"mtu"`
			Model                interface{} `json:"model"`
			Mac                  string      `json:"mac"`
			IP                   interface{} `json:"ip"`
			Bridge               string      `json:"bridge"`
		} `json:"nics"`
		Pcidevs []interface{} `json:"pcidevs"`
		Vfbs    []interface{} `json:"vfbs"`
		Vkbs    []interface{} `json:"vkbs"`
		Vtpms   []interface{} `json:"vtpms"`
	} `json:"config"`
	Domid int `json:"domid"`
}
