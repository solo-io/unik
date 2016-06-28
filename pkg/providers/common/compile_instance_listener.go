package common

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/instance-listener/bindata"
	"github.com/emc-advanced-dev/unik/pkg/compilers/rump"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"fmt"
)

func CompileInstanceListener(sourceDir, instanceListenerPrefix, dockerImage string, createImageFunc func(kernel, args string, mntPoints, bakedEnv []string, staticIpConfig string, noCleanup bool) (*types.RawImage, error)) (*types.RawImage, error) {
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

	staticAddr := os.Getenv("IL_ADDR")
	staticNetmask := os.Getenv("IL_NETMASK")
	staticGateway := os.Getenv("IL_GATEWAY")
	var staticIpConfig string
	if staticAddr != "" && staticNetmask != "" && staticGateway != "" {
		staticIpConfig = fmt.Sprintf("%s,%s,%s", staticAddr, staticNetmask, staticGateway)
	}

	params := types.CompileImageParams{
		SourcesDir: sourceDir,
		Args:       "-prefix " + instanceListenerPrefix,
		MntPoints:  []string{"/data"},
		StaticIpConfig: staticIpConfig,
	}
	rumpGoCompiler := &rump.RumpGoCompiler{
		RumCompilerBase: rump.RumCompilerBase{
			DockerImage: dockerImage,
			CreateImage: createImageFunc,
		},
	}
	return rumpGoCompiler.CompileRawImage(params)
}
