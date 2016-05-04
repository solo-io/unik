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
	ExtraConfig    ExtraConfig     `json:"ExtraConfig"`
}

func (image *Image) Copy() *Image {
	return &Image{
		Id:             image.Id,
		Name:           image.Name,
		DeviceMappings: image.DeviceMappings,
		SizeMb:         image.SizeMb,
		Infrastructure: image.Infrastructure,
		Created:        image.Created,
		ExtraConfig:    image.ExtraConfig,
	}
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
	ExtraConfig    ExtraConfig    `json:"ExtraConfig"`
}

func (instance *Instance) Copy() *Instance {
	return &Instance{
		Id:             instance.Id,
		ImageId:        instance.ImageId,
		Infrastructure: instance.Infrastructure,
		Name:           instance.Name,
		State:          instance.State,
		Created:        instance.Created,
		ExtraConfig:        instance.ExtraConfig,
	}
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

func (volume *Volume) Copy() *Volume{
	return &Volume{
		Id:             volume.Id,
		Name:           volume.Name,
		SizeMb:         volume.SizeMb,
		Attachment:     volume.Attachment,
		Infrastructure: volume.Infrastructure,
		Created:        volume.Created,
	}
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

//ExtraConfig exists for Compilers to specify special instructions to provider; should be used
//in the case that an Image/Instance/Volume should be run with non-default parameters
//(e.g. attach SATA controller instead of SCSI)
type ExtraConfig map[string]string

type RawImage struct {
	LocalImagePath string          `json:"LocalImagePath"`
	ExtraConfig    ExtraConfig     `json:"ExtraConfig"`
	DeviceMappings []DeviceMapping `json:"DeviceMappings"`
}

type RawVolume struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}
