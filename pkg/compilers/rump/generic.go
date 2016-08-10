package rump

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/emc-advanced-dev/pkg/errors"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"path/filepath"
)

func BuildBootableImage(kernel, cmdline string, usePartitionTables, noCleanup bool) (string, error) {
	directory, err := ioutil.TempDir("", "bootable-image-directory.")
	if err != nil {
		return "", err
	}
	if !noCleanup {
		defer os.RemoveAll(directory)
	}
	kernelBaseName := "program.bin"

	if err := unikos.CopyDir(filepath.Dir(kernel), directory); err != nil {
		return "", err
	}

	if err := unikos.CopyFile(kernel, path.Join(directory, kernelBaseName)); err != nil {
		return "", err
	}

	const contextDir = "/opt/vol/"
	cmds := []string{
		"-d", contextDir,
		"-p", kernelBaseName,
		"-a", cmdline,
		fmt.Sprintf("-part=%v", usePartitionTables),
	}
	binds := map[string]string{directory: contextDir, "/dev/": "/dev/"}

	if err := execContainer("boot-creator", cmds, binds, true, nil); err != nil {
		return "", err
	}

	resultFile, err := ioutil.TempFile("", "boot-creator-result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(directory, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}

func execContainer(imageName string, cmds []string, binds map[string]string, privileged bool, env map[string]string) error {
	container := unikutil.NewContainer(imageName).Privileged(privileged).WithVolumes(binds).WithEnvs(env)
	if err := container.Run(cmds...); err != nil {
		return errors.New("running container "+imageName, err)
	}
	return nil
}

type RumCompilerBase struct {
	DockerImage string
	CreateImage func(kernel, args string, mntPoints, bakedEnv []string, noCleanup bool) (*types.RawImage, error)
}

func (r *RumCompilerBase) runContainer(localFolder string, envPairs []string) error {
	env := make(map[string]string)
	for _, pair := range envPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return errors.New(pair+" is invaid string for env pair", nil)
		}
		env[split[0]] = split[1]
	}

	return unikutil.NewContainer(r.DockerImage).WithVolume(localFolder, "/opt/code").WithEnvs(env).Run()

}

func setRumpCmdLine(c rumpConfig, prog string, argv []string, addStub bool) rumpConfig {
	if addStub {
		stub := commandLine{
			Bin:  "stub",
			Argv: []string{},
		}
		c.Rc = append(c.Rc, stub)
	}
	progrc := commandLine{
		Bin:  "program",
		Argv: argv,
	}
	c.Rc = append(c.Rc, progrc)
	return c
}

var netRegEx = regexp.MustCompile("net[1-9]")
var envRegEx = regexp.MustCompile("\"env\":\\{(.*?)\\}")
var envRegEx2 = regexp.MustCompile("env[0-9]")

// rump special json
func toRumpJson(c rumpConfig) (string, error) {

	blk := c.Blk
	c.Blk = nil

	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	blks := ""
	for _, b := range blk {

		blkjson, err := json.Marshal(b)
		if err != nil {
			return "", err
		}
		blks += fmt.Sprintf("\"blk\": %s,", string(blkjson))
	}
	var jsonString string
	if len(blks) > 0 {

		jsonString = string(jsonConfig[:len(jsonConfig)-1]) + "," + blks[:len(blks)-1] + "}"

	} else {
		jsonString = string(jsonConfig)
	}

	jsonString = netRegEx.ReplaceAllString(jsonString, "net")

	jsonString = string(envRegEx.ReplaceAllString(jsonString, "$1"))

	jsonString = string(envRegEx2.ReplaceAllString(jsonString, "env"))

	return jsonString, nil

}
