package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type volumeslice []types.RawVolume

func (m *volumeslice) String() string {

	return fmt.Sprintf("%v", ([]types.RawVolume)(*m))
}

// The second method is Set(value string) error
func (m *volumeslice) Set(value string) error {
	volparts := strings.Split(value, ",")

	if (len(volparts) != 1) && (len(volparts) != 2) {
		return errors.New("bad format")
	}

	folder := volparts[0]

	var size int64
	if len(volparts) >= 2 {
		size, _ = strconv.ParseInt(volparts[1], 0, 64)
	}
	*m = append(*m, types.RawVolume{Path: folder, Size: size})

	return nil
}

func verifyPreConditions() {
	_, err := os.Stat("/dev/loop0")
	if os.IsNotExist(err) {
		log.Fatal("No loop device found. if running from docker use \"--privileged -v /dev/:/dev/\"")
	}
}
func main() {
	log.SetLevel(log.DebugLevel)

	var volumes volumeslice
	flag.Var(&volumes, "v", "volumes folder[,size]")
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")

	flag.Parse()

	if len(volumes) == 0 {
		log.Fatal("No volumes provided")
	}

	for i := range volumes {
		volumes[i].Path = path.Join(*buildcontextdir, volumes[i].Path)
	}

	imgFile := path.Join(*buildcontextdir, "vol.img")
	diskLabelGen := func(device string) unikos.Partitioner { return &unikos.DiskLabelPartioner{device} }

	verifyPreConditions()
	// rump so we use disklabel
	err := unikos.CreateVolumes(imgFile, []types.RawVolume(volumes), diskLabelGen)
	if err != nil {
		panic(err)
	}
}
