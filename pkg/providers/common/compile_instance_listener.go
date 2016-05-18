package common

import (
	"io/ioutil"
	"github.com/emc-advanced-dev/pkg/errors"
	"path/filepath"
	"github.com/emc-advanced-dev/unik/pkg/compilers/rump"
	"github.com/emc-advanced-dev/unik/instance-listener/bindata"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
)

func CompileInstanceListener(sourceDir, instanceListenerPrefix, dockerImage string, createImageFunc func(kernel, args string, mntPoints []string) (*types.RawImage, error)) (*types.RawImage, error) {
	mainData, err := bindata.Asset("instance-listener/main.go")
	if err != nil {
		return nil, errors.New("reading binary data of instance listener main", err)
	}
	if err := ioutil.WriteFile(filepath.Join(sourceDir, "main.go"), mainData, 0644); err != nil {
		return nil, errors.New("copying contents of instance listener main.go", err)
	}
	godepsData, err := bindata.Asset("instance-listener/Godeps/Godeps.json")
	if err != nil {
		return nil, errors.New("reading binary data of instance listener Godeps", err)
	}
	if err := os.MkdirAll(filepath.Join(sourceDir, "Godeps"), 0755); err != nil {
		return nil, errors.New("creating Godeps dir", err)
	}
	if err := ioutil.WriteFile(filepath.Join(sourceDir, "Godeps", "Godeps.json"), godepsData, 0644); err != nil {
		return nil, errors.New("copying contents of instance listener Godeps.json", err)
	}

	params := types.CompileImageParams{
		SourcesDir: sourceDir,
		Args: "-prefix "+instanceListenerPrefix,
		MntPoints: []string{"/data"},
	}
	rumpGoCompiler := &rump.RumpGoCompiler{
		DockerImage: dockerImage,
		CreateImage: createImageFunc,
	}
	return rumpGoCompiler.CompileRawImage(params)
}