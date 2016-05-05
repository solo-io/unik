package main

import (
	"os/exec"
	"os"
	"strings"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
	"path/filepath"
)

//expect project directory at /project_directory; mount w/ -v FOLDER:/project_directory
//output dir will be /project_directory
//output files to whatever is mounted to /project_directory
const (
	java_main_caller_dir = "/java-main-caller"
	project_directory = "/project_directory"
)

var buildImageTimeout = time.Minute * 10

func main() {
	out, _ := exec.Command("ls", "/").CombinedOutput()
	fmt.Println(strings.Split(string(out), "\n"))
	appInfo, err := wrapJavaApplication(java_main_caller_dir, project_directory)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to wrap java project Main Class", err)
		os.Exit(-1)
	}
	logrus.AddHook(&unikutil.AddTraceHook{true})
	fmt.Printf("read info from java project: %v\n", appInfo)

	fmt.Printf("runnning mvn clean")
	mvnClean := exec.Command("mvn", "clean")
	mvnClean.Dir = project_directory
	if err := mvnClean.Run(); err != nil {
		fmt.Printf("mvn clean failed, simply cleaning up %s/target", project_directory)
		os.RemoveAll(filepath.Join(project_directory, "target"))
	}

	fmt.Printf("running mvn package\n")

	mvnPackageCmd := exec.Command("mvn", "package")
	mvnPackageCmd.Dir = project_directory
	printCommand(mvnPackageCmd)
	if out, err := mvnPackageCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		os.Exit(-1)
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
		os.Exit(-1)
	}

	go func() {
		fmt.Println("capstain building")

		capstanCmd := exec.Command("capstan", "run", "-p", "qemu")
		capstanCmd.Dir = java_main_caller_dir
		capstanCmd.Stdout = os.Stdout
		capstanCmd.Stderr = os.Stderr
		printCommand(capstanCmd)
		if err := capstanCmd.Run(); err != nil {
			logrus.WithError(err).Error("captsain build failed")
			os.Exit(-1)
		}
	}()
	capstanImage := os.Getenv("HOME") + "/.capstan/instances/qemu/java-main-caller/disk.qcow2"

	select {
	case <-fileReady(capstanImage):
		fmt.Printf("image ready at %s\n", capstanImage)
		break
	case <-time.After(buildImageTimeout):
		logrus.Error("timed out waiting for capstan to finish building")
		os.Exit(-1)
	}

	fmt.Println("qemu-img converting (compatibility")
	convertToCompatibleCmd := exec.Command("qemu-img", "convert",
		"-f", "qcow2",
		"-O", "qcow2",
		"-o", "compat=0.10",
		capstanImage,
		project_directory + "/boot.qcow2")
	printCommand(convertToCompatibleCmd)
	if out, err := convertToCompatibleCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error(string(out))
		os.Exit(-1)
	}

	fmt.Println("file created at " + project_directory + "/boot.qcow2")
}

func fileReady(filename string) <-chan struct{} {
	closeChan := make(chan struct{})
	fmt.Printf("waiting for file to become ready...\n")
	go func() {
		count := 0
		for {
			if _, err := os.Stat(filename); err == nil {
				close(closeChan)
				return
			}
			//count every 5 sec
			if count%5 == 0 {
				fmt.Printf("waiting for file...%vs\n", count)
			}
			time.Sleep(time.Second * 1)
			count++
		}
	}()
	return closeChan
}


func printCommand(cmd *exec.Cmd) {
	fmt.Printf("running command from dir %s: %v\n", cmd.Dir, cmd.Args)
}
