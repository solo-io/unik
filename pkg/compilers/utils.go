package compilers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/emc-advanced-dev/pkg/errors"

	unikos "github.com/cf-unik/unik/pkg/os"
	unikutil "github.com/cf-unik/unik/pkg/util"
)

func BuildBootableImage(kernel, cmdline string, usePartitionTables, noCleanup bool) (string, error) {
	directory, err := ioutil.TempDir("", "bootable-image-directory.")
	if err != nil {
		return "", errors.New("creating tmpdir", err)
	}
	if !noCleanup {
		defer os.RemoveAll(directory)
	}
	kernelBaseName := "program.bin"

	if err := unikos.CopyDir(filepath.Dir(kernel), directory); err != nil {
		return "", errors.New("copying dir "+filepath.Dir(kernel)+" to "+directory, err)
	}

	if err := unikos.CopyFile(kernel, path.Join(directory, kernelBaseName)); err != nil {
		return "", errors.New("copying kernel "+kernel+" to "+kernelBaseName, err)
	}

	tmpResultFile, err := ioutil.TempFile(directory, "boot-creator-result.img.")
	if err != nil {
		return "", err
	}
	tmpResultFile.Close()

	const contextDir = "/opt/vol/"
	cmds := []string{
		"-d", contextDir,
		"-p", kernelBaseName,
		"-a", cmdline,
		"-o", filepath.Base(tmpResultFile.Name()),
		fmt.Sprintf("-part=%v", usePartitionTables),
	}
	binds := map[string]string{directory: contextDir, "/dev/": "/dev/"}

	if err := unikutil.NewContainer("boot-creator").Privileged(true).WithVolumes(binds).Run(cmds...); err != nil {
		return "", err
	}

	resultFile, err := ioutil.TempFile("", "boot-creator-result.img.")
	if err != nil {
		return "", err
	}
	resultFile.Close()

	if err := os.Rename(tmpResultFile.Name(), resultFile.Name()); err != nil {
		return "", errors.New("renaming "+tmpResultFile.Name()+" to "+resultFile.Name(), err)
	}
	return resultFile.Name(), nil
}
