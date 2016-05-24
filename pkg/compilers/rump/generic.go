package rump

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/emc-advanced-dev/pkg/errors"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

func BuildBootableImage(kernel, cmdline string) (string, error) {
	directory, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(directory)
	kernelBaseName := "program.bin"

	if err := unikos.CopyFile(kernel, path.Join(directory, kernelBaseName)); err != nil {
		return "", err
	}

	const contextDir = "/opt/vol/"
	cmds := []string{"-d", contextDir, "-p", kernelBaseName, "-a", cmdline}
	binds := []string{directory + ":" + contextDir, "/dev/:/dev/"}

	if err := execContainer("projectunik/boot-creator", cmds, binds, true, nil); err != nil {
		return "", err
	}

	resultFile, err := ioutil.TempFile(unikutil.UnikTmpDir(), "")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(path.Join(directory, "vol.img"), resultFile.Name()); err != nil {
		return "", err
	}

	return resultFile.Name(), nil
}

func execContainer(imageName string, cmds, binds []string, privileged bool, env map[string]string) error {
	dockerArgs := []string{"run", "--rm"}
	if privileged {
		dockerArgs = append(dockerArgs, "--privileged")
	}
	for _, bind := range binds {
		dockerArgs = append(dockerArgs, "-v", bind)
	}
	for key, val := range env {
		dockerArgs = append(dockerArgs, "-e", fmt.Sprintf("%s=%s", key, val))
	}
	dockerArgs = append(dockerArgs, imageName)
	dockerArgs = append(dockerArgs, cmds...)
	cmd := exec.Command("docker", dockerArgs...)
	unikutil.LogCommand(cmd, true)
	if err := cmd.Run(); err != nil {
		return errors.New("running container "+imageName, err)
	}
	return nil
}

func (r *RumpGoCompiler) runContainer(localFolder string, envPairs []string) error {
	env := make(map[string]string)
	for _, pair := range envPairs {
		split := strings.Split(pair, "=")
		if len(split) != 2 {
			return errors.New(pair+" is invaid string for env pair", nil)
		}
		env[split[0]] = split[1]
	}
	return execContainer(r.DockerImage, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false, env)
}

func (r *RumpGoCompiler) runAndBake(localFolder string, envPairs []string) error {
	if err := r.runContainer(localFolder, envPairs); err != nil {
		return err
	}
	// now we should program compiled in local folder. next step is to bake
	progFile := path.Join(localFolder, "program")

	if !unikos.IsExists(progFile) {
		return errors.New("No program found - compilation failed", nil)
	}

	return execContainer(r.BakeImageName, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false, nil)
}

func setRumpCmdLine(c rumpConfig, prog string, argv []string) rumpConfig {

	if argv == nil {
		argv = []string{}
	}

	pipe := "|"

	stub := commandLine{Bin: "stub",
		Argv: []string{},
	}
	progrc := commandLine{Bin: "program",
		Argv:    argv,
		Runmode: &pipe,
	}
	logger := commandLine{Bin: "logger",
		Argv: []string{},
	}

	c.Rc = append(c.Rc, stub, progrc, logger)
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
