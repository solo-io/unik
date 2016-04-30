package types

import (
	"fmt"
	"time"
)

type InstanceState string

const (
	InstanceState_Running    InstanceState = "running"
	InstanceState_Stopped    InstanceState = "stopped"
	InstanceState_Pending    InstanceState = "pending"
	InstanceState_Unknown    InstanceState = "unknown"
	InstanceState_Terminated InstanceState = "terminated"
)

type Infrastructure string

const (
	Infrastructure_AWS        Infrastructure = "AWS"
	Infrastructure_VSPHERE    Infrastructure = "VSPHERE"
	Infrastructure_VIRTUALBOX Infrastructure = "VIRTUALBOX"
)

type Image struct {
	Id             string          `json:"Id"`
	Name           string          `json:"Name"`
	DeviceMappings []DeviceMapping `json:"DeviceMappings"`
	SizeMb         int64           `json:"SizeMb"`
	Infrastructure Infrastructure  `json:"Infrastructure"`
	Created        time.Time       `json:"Created"`
}

func (image *Image) String() string {
	if image == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%-v", *image)
}

type Instance struct {
	Id             string         `json:"Id"`
	Name           string         `json:"Name"`
	State          InstanceState  `json:"State"`
	IpAddress      string         `json:"IpAddress"`
	ImageId        string         `json:"ImageId"`
	Infrastructure Infrastructure `json:"Infrastructure"`
	Created        time.Time      `json:"Created"`
}

func (instance *Instance) String() string {
	if instance == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", *instance)
}

type Volume struct {
	Id             string         `json:"Id"`
	Name           string         `json:"Name"`
	SizeMb         int64          `json:"SizeMb"`
	Attachment     string         `json:"Attachment"` //instanceId
	Infrastructure Infrastructure `json:"Infrastructure"`
	Created        time.Time      `json:"Created"`
}

func (volume *Volume) String() string {
	if volume == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%+v", *volume)
}

type DeviceMapping struct {
	MountPoint string `json:"MountPoint"`
	DeviceName string `json:"DeviceName"`
}

type RawImage struct {
	LocalImagePath string          `json:"LocalImagePath"`
	DeviceMappings []DeviceMapping `json:"DeviceMappings"`
}

type RawVolume struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}
