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
	"flag"
	"io/ioutil"
	"github.com/emc-advanced-dev/pkg/errors"
)

//expect project directory at /project_directory; mount w/ -v FOLDER:/project_directory
//output dir will be /project_directory
//output files to whatever is mounted to /project_directory
const (
	java_main_caller_udp_bootstrap_dir = "/java-main-caller-udp-bootstrap"
	java_main_caller_ec2_bootstrap_dir = "/java-main-caller-ec2-bootstrap"
	project_directory = "/project_directory"
)

var buildImageTimeout = time.Minute * 10

type appInfo struct {
	groupId       string
	artifactId    string
	version       string
	mainClassName string
}

func main() {
	useEc2Bootstrap := flag.Bool("ec2", false, "indicates whether to compile using the wrapper for ec2")
	groupId := flag.String("groupId", "", "groupid for jar file")
	artifactId := flag.String("artifactId", "", "artifactid for jar file")
	version := flag.String("version", "", "artifactid for jar file")
	mainClassName := flag.String("mainClassName", "", "mainClassName for jar file")
	jarPath := flag.String("jarName", "", "name of jar file (not path)")
	buildCmd := flag.String("buildCmd", "", "optional build command to build project (if not a jar)")
	args := flag.String("args", "", "arguments to kernel")
	flag.Parse()
	javaMainCallerDir := java_main_caller_udp_bootstrap_dir //use udp by default
	if *useEc2Bootstrap {
		javaMainCallerDir = java_main_caller_ec2_bootstrap_dir
	}
	out, _ := exec.Command("ls", "/").CombinedOutput()
	fmt.Println(strings.Split(string(out), "\n"))
	info := appInfo{
		groupId: *groupId,
		artifactId: *artifactId,
		version: *version,
		mainClassName: *mainClassName,
	}
	if err := wrapJavaApplication(info, javaMainCallerDir, project_directory); err != nil {
		logrus.WithError(err).Errorf("Failed to wrap java project Main Class", err)
		os.Exit(-1)
	}
	logrus.AddHook(&unikutil.AddTraceHook{true})

	if *buildCmd != "" {
		buildArgs := strings.Split(*buildCmd, " ")
		var params []string
		if len(buildArgs) > 1 {
			params = buildArgs[1:]
		}
		build := exec.Command(buildArgs[0], params...)
		build.Dir = project_directory
		build.Stdout = os.Stdout
		build.Stderr = os.Stderr
		printCommand(build)
		if err := build.Run(); err != nil {
			logrus.WithError(err).Error("failed running build command")
			os.Exit(-1)
		}
	}

	if _, err := os.Stat(filepath.Join(project_directory, *jarPath)); err != nil {
		logrus.WithError(err).Error("failed to stat "+filepath.Join(project_directory, *jarPath))
		os.Exit(-1)
	}

	mvnInstallCmd := exec.Command("mvn", "install:install-file",
		"-Dfile="+filepath.Join(project_directory, *jarPath),
		"-DgroupId=" + info.groupId,
		"-DartifactId=" + info.artifactId,
		"-Dversion=" + info.version,
		"-Dpackaging=jar")
	printCommand(mvnInstallCmd)
	if out, err := mvnInstallCmd.CombinedOutput(); err != nil {
		logrus.WithError(err).Error("failed running mvn install: "+ string(out))
		os.Exit(-1)
	}

	//add args to capstanfile
	if *args != "" {
		if err := addArgs(filepath.Join(javaMainCallerDir, "Capstanfile"), *args); err != nil {
			logrus.WithError(err).Error("adding capstan args failed")
			os.Exit(-1)
		}
	}

	go func() {
		fmt.Println("capstain building")

		capstanCmd := exec.Command("capstan", "run", "-p", "qemu")
		capstanCmd.Dir = javaMainCallerDir
		capstanCmd.Stdout = os.Stdout
		capstanCmd.Stderr = os.Stderr
		printCommand(capstanCmd)
		if err := capstanCmd.Run(); err != nil {
			logrus.WithError(err).Error("capstan build failed")
			os.Exit(-1)
		}
	}()
	capstanImage := filepath.Join(os.Getenv("HOME"), ".capstan", "instances", "qemu", javaMainCallerDir, "disk.qcow2")

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

func addArgs(capstanFile, args string) error {
	logrus.Infof("adding %s to args", args)
	capstanBytes, err := ioutil.ReadFile(capstanFile)
	if err != nil {
		return errors.New("reading capstanfile", err)
	}
	capstanContents := strings.Replace(string(capstanBytes), "cmdline: /java.so -jar /program.jar", fmt.Sprintf("cmdline: /java.so -jar /program.jar %s", args), -1)

	err = ioutil.WriteFile(capstanFile, []byte(capstanContents), 0666)
	if err != nil {
		return errors.New("writing capstanfile", err)
	}
	return nil
}

func wrapJavaApplication(info appInfo, javaWrapperDir, appSourceDir string) error {
	wrapperPomBytes, err := ioutil.ReadFile(javaWrapperDir + "/pom.xml")
	if err != nil {
		return errors.New("reading app pom bytes", err)
	}
	wrapperPomContents := strings.Replace(string(wrapperPomBytes), "REPLACE_WITH_GROUPID", info.groupId, -1)
	wrapperPomContents = strings.Replace(wrapperPomContents, "REPLACE_WITH_ARTIFACTID", info.artifactId, -1)
	wrapperPomContents = strings.Replace(wrapperPomContents, "REPLACE_WITH_VERSION", info.version, -1)

	err = ioutil.WriteFile(javaWrapperDir + "/pom.xml", []byte(wrapperPomContents), 0666)
	if err != nil {
		return errors.New("writing pom.xml", err)
	}

	wrapperMainContentBytes, err := ioutil.ReadFile(javaWrapperDir + "/src/main/java/com/emc/wrapper/Wrapper.java")
	if err != nil {
		return errors.New("reading java pom bytes", err)
	}
	wrapperMainContents := strings.Replace(string(wrapperMainContentBytes), "REPLACE_WITH_MAIN_CLASS", info.mainClassName, -1)

	err = ioutil.WriteFile(javaWrapperDir + "/src/main/java/com/emc/wrapper/Wrapper.java", []byte(wrapperMainContents), 0666)
	if err != nil {
		return errors.New("writing Wrapper class around app class", err)
	}

	return nil
}