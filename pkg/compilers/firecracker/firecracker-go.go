package firecracker

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/unik/pkg/compilers"
	"github.com/solo-io/unik/pkg/types"
	unikutil "github.com/solo-io/unik/pkg/util"
)

type FirecrackerCompiler struct{}

func (f *FirecrackerCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	sourcesDir := params.SourcesDir

	// run dep ensure and go build
	if err := unikutil.NewContainer("compilers-firecracker").Privileged(true).WithVolume(sourcesDir, "/opt/code").Run(); err != nil {
		return nil, err
	}
	res := &types.RawImage{}
	localImageFile, err := f.getImagefile(sourcesDir)
	if err != nil {
		logrus.Errorf("error getting local image file name")
	}
	res.LocalImagePath = localImageFile
	res.StageSpec.ImageFormat = types.ImageFormat_RAW
	res.RunSpec.DefaultInstanceMemory = 256
	return res, nil
}

func (f *FirecrackerCompiler) getImagefile(directory string) (string, error) {

	rootfs := filepath.Join(directory, "rootfs")

	_, err := os.Stat(rootfs)
	return rootfs, err
}

func (f *FirecrackerCompiler) Usage() *compilers.CompilerUsage {
	return nil
}
