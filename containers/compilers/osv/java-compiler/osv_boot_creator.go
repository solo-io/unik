package main

import (
	"os/exec"
	"os"
	"strings"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
)

//expect project directory at /project_directory; mount w/ -v FOLDER:/project_directory
//output dir will be /project_directory
//output files to whatever is mounted to /project_directory
const (
	java_main_caller_dir = "/java-main-caller"
	project_directory = "/project_directory"
)

func main() {
	out, _ := exec.Command("ls", "/").CombinedOutput()
	fmt.Println(strings.Split(string(out), "\n"))
	appInfo, err := wrapJavaApplication(java_main_caller_dir, project_directory)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to wrap java project Main Class", err)
		for {}
	}
	logrus.AddHook(&unikutil.AddTraceHook{true})
	fmt.Printf("read info from java project: %v", appInfo)
	fmt.Printf("running mvn package")

	mvnPackageCmd := exec.Command("mvn", "package")
	mvnPackageCmd.Dir = project_directory
	printCommand(mvnPackageCmd)
	if out, err := mvnPackageCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	mvnInstallCmd := exec.Command("mvn", "install:install-file",
		"-Dfile=target/" + appInfo.ArtifactId + "-" + appInfo.Version + "-jar-with-dependencies.jar",
		"-DgroupId=" + appInfo.GroupId,
		"-DartifactId=" + appInfo.ArtifactId,
		"-Dversion=" + appInfo.Version,
		"-Dpackaging=jar")
	mvnInstallCmd.Dir = project_directory
	printCommand(mvnInstallCmd)
	if out, err := mvnInstallCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	go func() {
		fmt.Println("capstain building")

		capstanCmd := exec.Command("capstan", "run", "-p", "qemu")
		capstanCmd.Dir = java_main_caller_dir
		printCommand(capstanCmd)
		if out, err := capstanCmd.CombinedOutput(); err != nil {
			logrus.WithError(err).Error(string(out))
			for {}
		}
	}()
	capstanImage := os.Getenv("HOME") + "/.capstan/instances/qe	out, _ := exec.Command("ls", "/").CombinedOutput()
	fmt.Println(strings.Split(string(out), "\n"))
	appInfo, err := wrapJavaApplication(java_main_caller_dir, project_directory)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to wrap java project Main Class", err)
		for {}
	}
	logrus.AddHook(&unikutil.AddTraceHook{true})
	fmt.Printf("read info from java project: %v", appInfo)
	fmt.Printf("running mvn package")

	mvnPackageCmd := exec.Command("mvn", "package")
	mvnPackageCmd.Dir = project_directory
	printCommand(mvnPackageCmd)
	if out, err := mvnPackageCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	mvnInstallCmd := exec.Command("mvn", "install:install-file",
		"-Dfile=target/" + appInfo.ArtifactId + "-" + appInfo.Version + "-jar-with-dependencies.jar",
		"-DgroupId=" + appInfo.GroupId,
		"-DartifactId=" + appInfo.ArtifactId,
		"-Dversion=" + appInfo.Version,
		"-Dpackaging=jar")
	mvnInstallCmd.Dir = project_directory
	printCommand(mvnInstallCmd)
	if out, err := mvnInstallCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	go func() {
		fmt.Println("capstain building")

		capstanCmd := exec.Command("capstan", "run", "-p", "qemu")
		capstanCmd.Dir = java_main_caller_dir
		printCommand(capstanCmd)
		if out, err := capstanCmd.CombinedOutput(); err != nil {
			logrus.WithError(err).Error(string(out))
			for {}
		}
	}()
	capstanImage := os.Getenv("HOME") + "/.capstan/instances/qemu/java-main-caller/disk.qcow2"

	select {
	case <-fileReady(capstanImage):
		fmt.Printf("image ready at %s\n", capstanImage)
		break
	case <-time.After(time.Second * 120):
		logrus.Error("capstan never finished building")
		for {}
	}

	fmt.Println("qemu-img creating")
	convertToRawCmd := exec.Command("qemu-img", "convert",
		"-O", "vmdk",
		capstanImage,
		project_directory + "/boot.vmdk")
	printCommand(convertToRawCmd)
	if out, err := convertToRawCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	fmt.Println("file created at " + project_directory + "/boot.vmdk")
	mu/java-main-caller/disk.qcow2"

	select {
	case <-fileReady(capstanImage):
		fmt.Printf("image ready at %s\n", capstanImage)
		break
	case <-time.After(time.Second * 120):
		logrus.Error("capstan never finished building")
		for {}
	}

	fmt.Println("qemu-img creating")
	convertToRawCmd := exec.Command("qemu-img", "convert",
		"-O", "vmdk",
		capstanImage,
		project_directory + "/boot.vmdk")
	printCommand(convertToRawCmd)
	if out, err := convertToRawCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		for {}
	}

	fmt.Println("file created at " + project_directory + "/boot.vmdk")
}

func fileReady(filename string) <-chan struct{} {
	closeChan := make(chan struct{})
	go func() {
		count := 0
		for {
			fmt.Printf("waiting for file...%s", count)
			if _, err := os.Stat(filename); err == nil {
				close(closeChan)
				return
			}
			time.Sleep(time.Second)
		}
	}()
	return closeChan
}

func printCommand(cmd *exec.Cmd) {
	fmt.Printf("running command from dir %s: %v\n", cmd.Dir, cmd.Args)
}
