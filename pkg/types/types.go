package types

import "time"

type InstanceState string

const (
	InstanceState_Running     InstanceState = "InstanceState_Running"
	InstanceState_Stopped     InstanceState = "InstanceState_Stopped"
	InstanceState_Terminating InstanceState = "InstanceState_Terminating"
	InstanceState_Pending     InstanceState = "InstanceState_Pending"
	InstanceState_Unknown     InstanceState = "InstanceState_Unknown"
)

type Infrastructure string

const (
	Infrastructure_AWS Infrastructure = "Infrastructure_AWS"
	Infrastructure_VSPHERE Infrastructure = "Infrastructure_VSPHERE"
	Infrastructure_VIRTUALBOX Infrastructure = "Infrastructure_VIRTUALBOX"
)

type Image struct {
	Id             string           `json:"Id"`
	Name           string           `json:"Name"`
	DeviceMappings []*DeviceMapping `json:"DeviceMappings"`
	SizeMb         int64            `json:"SizeMb"`
	Infrastructure Infrastructure   `json:"Infrastructure"`
	Created	       time.Time	`json:"Created"`
}

type Instance struct {
	Id             string        `json:"Id"`
	Name           string        `json:"Name"`
	State          InstanceState `json:"State"`
	IpAddress      string        `json:"IpAddress"`
	ImageId        string        `json:"ImageId"`
	Infrastructure string        `json:"Infrastructure"`
	Created	       time.Time	`json:"Created"`
}

type Volume struct {
	Id             string `json:"Id"`
	Name           string `json:"Name"`
	SizeMb         int64  `json:"SizeMb"`
	Attachment     string `json:"Attachment"` //instanceId
	Infrastructure string `json:"Infrastructure"`
	Created	       time.Time	`json:"Created"`
}

type DeviceMapping struct {
	MountPoint string `json:"MountPoint"`
	DeviceName string `json:"DeviceName"`
}

type RawImage struct {
	LocalImagePath string           `json:"LocalImagePath"`
	DeviceMappings []*DeviceMapping `json:"DeviceMappings"`
}

type RawVolume struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}
