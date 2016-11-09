// +build !linux

package os

func Mount(device BlockDevice) (mntpoint string, err error) {
	panic("Not supported")
}
func MountDevice(device string) (mntpoint string, err error) {
	panic("Not supported")
}

func Umount(point string) error {
	panic("Not supported")
}

type MsDosPartioner struct {
	Device string
}

func (m *MsDosPartioner) MakeTable() error {
	panic("Not supported")
}

func (m *MsDosPartioner) MakePart(partType string, start, size DiskSize) error {
	panic("Not supported")
	return nil
}
func (m *MsDosPartioner) MakePartTillEnd(partType string, start DiskSize) error {
	panic("Not supported")
	return nil
}

func (m *MsDosPartioner) Makebootable(partnum int) error {
	panic("Not supported")
	return nil
}

type DiskLabelPartioner struct {
	Device string
}

func (m *DiskLabelPartioner) MakeTable() error {
	panic("Not supported")
	return nil
}

func (m *DiskLabelPartioner) MakePart(partType string, start, size DiskSize) error {
	panic("Not supported")
	return nil
}

func ListParts(device BlockDevice) ([]Part, error) {
	panic("Not supported")
	return nil, nil
}

type PartedPart struct {
	Device BlockDevice
}

func (p *PartedPart) Size() DiskSize {
	panic("Not supported")
	return Bytes(0)
}
func (p *PartedPart) Offset() DiskSize {
	panic("Not supported")
	return Bytes(0)
}

func (p *PartedPart) Acquire() (BlockDevice, error) {

	panic("Not supported")
	return "", nil
}

func (p *PartedPart) Release() error {
	panic("Not supported")
	return nil
}

func (p *PartedPart) Get() BlockDevice {
	panic("Not supported")
	return ""
}

type DeviceMapperDevice struct {
	DeviceName string
}

func NewDevice(start, size Sectors, origDevice BlockDevice, deivceName string) Resource {
	panic("Not supported")
	return nil
}

func (p *DeviceMapperDevice) Size() DiskSize {
	panic("Not supported")
	return Bytes(0)
}
func (p *DeviceMapperDevice) Offset() DiskSize {
	panic("Not supported")
	return Bytes(0)
}

func (p *DeviceMapperDevice) Acquire() (BlockDevice, error) {

	panic("Not supported")
	return "", nil
}

func (p *DeviceMapperDevice) Release() error {

	panic("Not supported")
	return nil
}

func (p *DeviceMapperDevice) Get() BlockDevice {

	panic("Not supported")
	return ""
}

type LoDevice struct {
}

func NewLoDevice(device string) Resource {

	panic("Not supported")
	return nil
}

func (p *LoDevice) Acquire() (BlockDevice, error) {

	panic("Not supported")
	return "", nil
}

func (p *LoDevice) Release() error {

	panic("Not supported")
	return nil
}
