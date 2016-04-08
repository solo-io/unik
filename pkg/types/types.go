package types

type InstanceState string

const (
	InstanceState_Running     = "InstanceState_Running"
	InstanceState_Stopped     = "InstanceState_Stopped"
	InstanceState_Terminating = "InstanceState_Terminating"
	InstanceState_Pending     = "InstanceState_Pending"
)

type Instance struct {
	Name           string        `json:"Name"`
	Id             string        `json:"Id"`
	State          InstanceState `json:"State"`
	ImageId        string        `json:"ImageId"`
	Infrastructure string        `json:"Infrastructure"`
}

type Image struct {
	Name           string           `json:"Name"`
	Id             string           `json:"Id"`
	DeviceMappings []*DeviceMapping `json:"DeviceMappings"`
	SizeMb         int64            `json:"SizeMb"`
	Infrastructure string           `json:"Infrastructure"`
}

type Volume struct {
	Name           string `json:"Name"`
	Id             string `json:"Id"`
	SizeMb         int64  `json:"SizeMb"`
	Attachment     string `json:"Attachment"` //instanceId
	Infrastructure string `json:"Infrastructure"`
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
