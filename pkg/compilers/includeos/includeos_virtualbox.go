package includeos

import (
	goerrors "errors"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

type IncludeosVirtualboxCompiler struct{}

func (i *IncludeosVirtualboxCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	sourcesDir := params.SourcesDir
	env := make(map[string]string)
	if err := unikutil.NewContainer("compilers-includeos-cpp-hw").WithVolume(sourcesDir, "/opt/code").WithEnvs(env).Run(); err != nil {
		return nil, err
	}

	res := &types.RawImage{}
	localImageFile, err := i.findFirstImageFile(sourcesDir)
	if err != nil {
		return nil, errors.New("error getting local image file name", err)
	}
	res.LocalImagePath = path.Join(sourcesDir, localImageFile)
	res.StageSpec.ImageFormat = types.ImageFormat_RAW
	res.RunSpec.StorageDriver = types.StorageDriver_IDE
	res.RunSpec.DefaultInstanceMemory = 256
	return res, nil
}

func (i *IncludeosVirtualboxCompiler) findFirstImageFile(directory string) (string, error) {
	dir, err := os.Open(directory)
	if err != nil {
		return "", errors.New("could not open dir", err)
	}
	defer dir.Close()
	files, err := dir.Readdir(-1)
	if err != nil {
		return "", errors.New("could not read dir", err)
	}
	for _, file := range files {
		logrus.Debugf("searching for .img file: %v", file.Name())
		if file.Mode().IsRegular() {
			if filepath.Ext(file.Name()) == ".img" {
				return file.Name(), nil
			}
		}
	}
	return "", errors.New("no image file found", goerrors.New("end of dir"))
}

func (r *IncludeosVirtualboxCompiler) Usage() *compilers.CompilerUsage {
	return nil
}
