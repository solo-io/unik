package main

import (
	"os/exec"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/Sirupsen/logrus"
	"os"
	"strings"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
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
	logrus.Info(strings.Split(string(out), "\n"))
	appInfo, err := wrapJavaApplication(java_main_caller_dir, project_directory)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to wrap java project Main Class", err)
		for {}
	}
	logrus.AddHook(&unikutil.AddTraceHook{true})
	logrus.Infof("read info from java project: %v", appInfo)
	logrus.Infof("running mvn package")
	mvnPackageCmd := exec.Command("mvn", "package")
	mvnPackageCmd.Dir = project_directory
	util.LogCommand(mvnPackageCmd, false)
	if err := mvnPackageCmd.Run(); err != nil {
		logrus.Error(err)
		for {}
	}
	mvnInstallCmd := exec.Command("mvn", "install",
		"-Dfile=target/"+appInfo.ArtifactId+"-"+appInfo.Version+"-jar-with-dependencies.jar",
		"-DgroupId="+appInfo.GroupId,
		"-DartifactId="+appInfo.ArtifactId,
		"-Dversion="+appInfo.Version,
		"-Dpackaging=jar")
	mvnInstallCmd.Dir = project_directory
	util.LogCommand(mvnInstallCmd, false)
	if err := mvnInstallCmd.Run(); err != nil {
		logrus.Error(err)
		for {}
	}
	logrus.Infof("running mvn install: %v", mvnInstallCmd.Args)
	logrus.Info("capstain building")
	capstanCmd := exec.Command("capstan", "build", "-p", "qemu", "boot")
	capstanCmd.Dir = java_main_caller_dir
	util.LogCommand(capstanCmd, false)
	if err := capstanCmd.Run(); err != nil {
		logrus.Error(err)
		for {}
	}
	logrus.Info("qemu-img creating")
	convertToRawCmd := exec.Command("qemu-img", "create",
		"-f", "qcow2",
		"-o", "backing_file="+os.Getenv("HOME")+"/.capstan/repository/boot/boot.qemu",
		"-F", "raw",
		project_directory+"/program.img")
	util.LogCommand(convertToRawCmd, false)
	if err := convertToRawCmd.Run(); err != nil {
		logrus.Error(err)
		for {}
	}
	logrus.Info("file created at "+project_directory+"/program.img")
}
