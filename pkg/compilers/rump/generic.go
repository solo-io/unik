package rump

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"os/exec"
	"strings"
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
