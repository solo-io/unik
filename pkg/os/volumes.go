package os

import (
	"io/ioutil"
	"os"
	"path"
	"text/template"

	"github.com/emc-advanced-dev/pkg/errors"

	log "github.com/Sirupsen/logrus"
)

type RawVolume struct {
	Path string `json:"Path"`
	Size int64  `json:"Size"`
}

const GrubTemplate = `default=0
fallback=1
timeout=1
hiddenmenu

title Unik
root {{.RootDrive}}
kernel /boot/program.bin {{.CommandLine}}
`

const DeviceMapFile = `(hd0) {{.GrubDevice}}
`

const ProgramName = "program.bin"

func createSparseFile(filename string, size DiskSize) error {
	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = fd.Seek(int64(size.ToBytes())-1, 0)
	if err != nil {
		return err
	}
	_, err = fd.Write([]byte{0})
	if err != nil {
		return err
	}
	return nil
}

func CreateBootImageWithSize(rootFile string, size DiskSize, progPath, staticFilesDir, commandline string, usePartitionTables bool) error {
	err := createSparseFile(rootFile, size)
	if err != nil {
		return err
	}
	log.WithFields(log.Fields{"imgFile": rootFile, "size": size.ToPartedFormat()}).Debug("created sparse file")

	if usePartitionTables {
		return CreateBootImageOnFile(rootFile, progPath, staticFilesDir, commandline)
	}
	return CreateBootImageOnFilePvGrub(rootFile, progPath, staticFilesDir, commandline)
}

func CreateBootImageOnFile(rootFile string, progPath, staticFilesDir, commandline string) error {

	log.WithFields(log.Fields{"imgFile": rootFile}).Debug("attaching sparse file")
	rootLo := NewLoDevice(rootFile)
	rootLodName, err := rootLo.Acquire()
	if err != nil {
		return err
	}
	defer rootLo.Release()

	log.Debug("device mapping to 'hda'")

	// use device mapper to rename the lo device to something that grub likes more.
	// like hda!
	grubDiskName := "hda"

	devTmp, err := ioutil.TempDir("/dev", "unik-tmp")
	if err != nil {
		return err
	}
	defer os.RemoveAll(devTmp)

	rootDeviceName := path.Join(devTmp, grubDiskName)
	if err := os.Link(rootLodName.Name(), rootDeviceName); err != nil {
		return err
	}

	log.Debug("partitioning")

	p := &MsDosPartioner{rootLodName.Name()}
	if err := p.MakeTable(); err != nil {
		return err
	}
	if err := p.MakePartTillEnd("primary", MegaBytes(2)); err != nil {
		return err
	}
	// make the partition just created bootable
	if err := p.Makebootable(1); err != nil {
		return err
	}
	parts, err := ListParts(rootLodName)
	if err != nil {
		return err
	}

	if len(parts) < 1 {
		return errors.New("No parts created", nil)
	}

	part := parts[0]

	// get the block device
	bootDevice, err := part.Acquire()
	if err != nil {
		return err
	}
	defer part.Release()

	firstPart := rootDeviceName + "1"
	if err := os.Link(bootDevice.Name(), firstPart); err != nil {
		return err
	}

	bootLabel := "boot"
	// format the device and mount and copy
	err = RunLogCommand("mkfs", "-L", bootLabel, "-I", "128", "-t", "ext2", bootDevice.Name())
	if err != nil {
		return err
	}

	mntPoint, err := MountDevice(firstPart)
	if err != nil {
		return err
	}
	defer Umount(mntPoint)

	if err := PrepareGrub(mntPoint, rootDeviceName, progPath, staticFilesDir, commandline); err != nil {
		return err
	}

	err = RunLogCommand("grub-install", "--no-floppy", "--root-directory="+mntPoint, rootDeviceName)
	if err != nil {
		return err
	}

	return nil
}

func PrepareGrub(folder, rootDeviceName, kernel, staticFilesDir, commandline string) error {
	grubPath := path.Join(folder, "boot", "grub")
	kernelDst := path.Join(folder, "boot", ProgramName)

	os.MkdirAll(grubPath, 0755)

	log.WithFields(log.Fields{"src": staticFilesDir, "dst": folder}).Debug("copying all files")
	if err := CopyDir(staticFilesDir, folder); err != nil {
		return err
	}

	// copy program.bin.. skip that for now
	log.WithFields(log.Fields{"src": kernel, "dst": kernelDst}).Debug("copying file")
	if err := CopyFile(kernel, kernelDst); err != nil {
		return err
	}

	if err := writeBootTemplate(path.Join(grubPath, "menu.lst"), "(hd0,0)", commandline); err != nil {
		return err
	}

	if err := writeBootTemplate(path.Join(grubPath, "grub.conf"), "(hd0,0)", commandline); err != nil {
		return err
	}

	if err := writeDeviceMap(path.Join(grubPath, "device.map"), rootDeviceName); err != nil {
		return err
	}
	return nil
}

func CreateBootImageOnFilePvGrub(rootFile string, progPath, staticFilesDir, commandline string) error {
	log.WithFields(log.Fields{"imgFile": rootFile}).Debug("attaching sparse file")
	rootLo := NewLoDevice(rootFile)
	bootDevice, err := rootLo.Acquire()
	if err != nil {
		return err
	}
	defer rootLo.Release()

	bootLabel := "boot"
	// format the device and mount and copy
	err = RunLogCommand("mkfs", "-L", bootLabel, "-I", "128", "-t", "ext2", bootDevice.Name())
	if err != nil {
		return err
	}

	mntPoint, err := Mount(bootDevice)
	if err != nil {
		return err
	}
	defer Umount(mntPoint)

	if err := PreparePVGrub(mntPoint, "sda1", progPath, staticFilesDir, commandline); err != nil {
		return err
	}

	return nil
}

func PreparePVGrub(folder, rootDeviceName, kernel, staticFilesDir, commandline string) error {
	grubPath := path.Join(folder, "boot", "grub")
	kernelDst := path.Join(folder, "boot", ProgramName)

	os.MkdirAll(grubPath, 0755)

	log.WithFields(log.Fields{"src": staticFilesDir, "dst": folder}).Debug("copying all files")
	if err := CopyDir(staticFilesDir, folder); err != nil {
		return err
	}

	// copy program.bin.. skip that for now
	log.WithFields(log.Fields{"src": kernel, "dst": kernelDst}).Debug("copying file")
	if err := CopyFile(kernel, kernelDst); err != nil {
		return err
	}

	if err := writeBootTemplate(path.Join(grubPath, "menu.lst"), "(hd0)", commandline); err != nil {
		return err
	}

	if err := writeBootTemplate(path.Join(grubPath, "grub.conf"), "(hd0)", commandline); err != nil {
		return err
	}

	if err := writeDeviceMap(path.Join(grubPath, "device.map"), rootDeviceName); err != nil {
		return err
	}
	return nil
}

func writeDeviceMap(fname, rootDevice string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.Must(template.New("devicemap").Parse(DeviceMapFile))

	log.WithFields(log.Fields{"device": rootDevice, "file": fname}).Debug("Writing device map")
	if err := t.Execute(f, struct {
		GrubDevice string
	}{rootDevice}); err != nil {
		return err
	}

	return nil
}
func writeBootTemplate(fname, rootDrive, commandline string) error {
	log.WithFields(log.Fields{"fname": fname, "rootDrive": rootDrive, "commandline": commandline}).Debug("writing boot template")

	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	t := template.Must(template.New("grub").Parse(GrubTemplate))

	if err := t.Execute(f, struct {
		RootDrive   string
		CommandLine string
	}{rootDrive, commandline}); err != nil {
		return err
	}

	return nil

}

func formatDeviceAndCopyContents(folder string, volType string, dev BlockDevice) error {
	var err error
	switch volType {
	case "fat":
		err = RunLogCommand("mkfs.fat", dev.Name())
	case "ext2":
		fallthrough
	case "":
		err = RunLogCommand("mkfs", "-I", "128", "-t", "ext2", dev.Name())
	default:
		return errors.New("Unknown fs type", nil)
	}

	if err != nil {
		return err
	}

	mntPoint, err := Mount(dev)
	if err != nil {
		return err
	}
	defer Umount(mntPoint)

	if err := CopyDir(folder, mntPoint); err != nil {
		return err
	}
	return nil
}

func CreateSingleVolume(rootFile string, volType string, folder RawVolume) error {
	ext2Overhead := MegaBytes(2).ToBytes()

	size := folder.Size

	if size == 0 {
		var err error
		size, err = GetDirSize(folder.Path)
		if err != nil {
			return err
		}
	}

	// take a spare sizde and down to sector size
	size = (SectorSize + size + size/10 + int64(ext2Overhead))
	size &^= (SectorSize - 1)
	// 10% buffer.. aligned to 512
	sizeVolume := Bytes(size)

	if _, err := ToSectors(Bytes(size)); err != nil {
		return err
	}

	if err := createSparseFile(rootFile, sizeVolume); err != nil {
		return err
	}

	return CopyToImgFile(folder.Path, volType, rootFile)
}

func CopyToImgFile(folder, volType string, imgfile string) error {
	imgLo := NewLoDevice(imgfile)
	imgLodName, err := imgLo.Acquire()
	if err != nil {
		return err
	}
	defer imgLo.Release()

	return formatDeviceAndCopyContents(folder, volType, imgLodName)

}

func copyToPart(folder string, volType string, part Resource) error {
	imgLodName, err := part.Acquire()
	if err != nil {
		return err
	}
	defer part.Release()
	return formatDeviceAndCopyContents(folder, volType, imgLodName)
}

func CreateVolumes(imgFile string, volType string, volumes []RawVolume, newPartitioner func(device string) Partitioner) error {
	if len(volumes) == 0 {
		return nil
	}

	var sizes []Bytes

	ext2Overhead := MegaBytes(2).ToBytes()
	firstPartOffest := MegaBytes(2).ToBytes()
	var totalSize Bytes = 0

	log.Debug("Calculating sizes")

	for _, v := range volumes {
		if v.Size == 0 {
			cursize, err := GetDirSize(v.Path)
			if err != nil {
				return err
			}
			sizes = append(sizes, Bytes(cursize)+ext2Overhead)
		} else {
			sizes = append(sizes, Bytes(v.Size))
		}
		totalSize += sizes[len(sizes)-1]
	}
	sizeDrive := Bytes((SectorSize + totalSize + totalSize/10) &^ (SectorSize - 1))
	sizeDrive += MegaBytes(4).ToBytes()

	log.WithFields(log.Fields{"imgFile": imgFile, "size": totalSize.ToPartedFormat()}).Debug("Creating image file")
	err := createSparseFile(imgFile, sizeDrive)
	if err != nil {
		return err
	}

	imgLo := NewLoDevice(imgFile)
	imgLodName, err := imgLo.Acquire()
	if err != nil {
		return err
	}
	defer imgLo.Release()

	p := newPartitioner(imgLodName.Name())

	p.MakeTable()
	var start Bytes = firstPartOffest
	for _, curSize := range sizes {
		end := start + curSize
		log.WithFields(log.Fields{"start": start, "end": end}).Debug("Creating partition")

		err := p.MakePart(toPartedVolType(volType), start, end)
		if err != nil {
			return err
		}
		curParts, err := ListParts(imgLodName)
		if err != nil {
			return err
		}
		start = curParts[len(curParts)-1].Offset().ToBytes() + curParts[len(curParts)-1].Size().ToBytes()
	}

	parts, err := ListParts(imgLodName)

	if len(parts) != len(volumes) {
		return errors.New("Not enough parts created!", nil)
	}

	log.WithFields(log.Fields{"parts": parts, "volsize": sizes}).Debug("Creating volumes")
	for i, v := range volumes {

		if err := copyToPart(v.Path, volType, parts[i]); err != nil {
			return err
		}
	}

	return nil
}

func toPartedVolType(volType string) string {
	switch volType {
	case "fat":
		return "fat32"
	default:
		return volType
	}
}
