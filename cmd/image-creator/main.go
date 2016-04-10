// +build linux

package main

import (
	"errors"
	"flag"
	"fmt"
	"path"
	"strconv"
	"strings"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

type volumeslice []types.RawVolume

func (m volumeslice) String() string {

	return fmt.Sprintf("%v", ([]types.RawVolume)(m))
}

// The second method is Set(value string) error
func (m volumeslice) Set(value string) error {
	volparts := strings.Split(value, ",")

	if (len(volparts) != 1) && (len(volparts) != 2) {
		return errors.New("bad format")

	}
	folder := volparts[0]

	var size int64
	if len(volparts) >= 2 {
		size, _ = strconv.ParseInt(volparts[1], 0, 64)
	}
	m = append(m, types.RawVolume{Path: folder, Size: size})

	return nil
}

func main() {
	var volumes volumeslice
	flag.Var(volumes, "v", "volumes localdir:remotedir")
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")

	for _, r := range volumes {
		r.Path = path.Join(*buildcontextdir, r.Path)
	}

	imgFile := ""
	diskLabelGen := func(device string) unikos.Partitioner { return &unikos.DiskLabelPartioner{device} }
	// rump so we use disklabel
	err := unikos.CreateVolumes(imgFile, []types.RawVolume(volumes), diskLabelGen)
	if err != nil {
		panic(err)
	}
}
