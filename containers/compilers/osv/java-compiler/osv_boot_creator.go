package main

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//expect project directory at /project_directory; mount w/ -v FOLDER:/project_directory
//output dir will be /project_directory
//output files to whatever is mounted to /project_directory
const (
	project_directory = "/project_directory"
)

var buildImageTimeout = time.Minute * 10

func main() {
	useEc2Bootstrap := flag.Bool("ec2", false, "indicates whether to compile using the wrapper for ec2")
	artifactName := flag.String("artifactName", "", "name of jar or war file (not path)")
	buildCmd := flag.String("buildCmd", "", "optional build command to build project (if not a jar)")
	properties := flag.String("properties", "", "-D properties passed to jar apps")
	args := flag.String("args", "", "arguments to kernel")
	flag.Parse()

	if *buildCmd != "" {
		logrus.WithField("cmd", *buildCmd).Info("running user specified build command")
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

	artifactFile := filepath.Join(project_directory, *artifactName)
	if _, err := os.Stat(artifactFile); err != nil {
		logrus.WithError(err).Error("failed to stat " + filepath.Join(project_directory, *artifactName) + "; is artifact_file set correctly?")
		logrus.Info("listing project files for debug purposes:")
		listProjectFiles := exec.Command("find", project_directory)
		listProjectFiles.Stdout = os.Stdout
		listProjectFiles.Stderr = os.Stderr
		listProjectFiles.Run()
		os.Exit(-1)
	}

	propertyArr := strings.Split(*properties, " ")
	proprtiesStr := ""
	for _, prop := range propertyArr {
		//format key=val
		proprtiesStr = fmt.Sprintf("-D%s %s", prop, proprtiesStr)
	}
	argsStr := ""
	if *useEc2Bootstrap {
		argsStr += "-bootstrapType=ec2 "
	} else {
		argsStr += "-bootstrapType=udp "
	}
	if *args != "" {
		argsStr += fmt.Sprintf("-appArgs=%s ", strings.Join(strings.Split(*args, " "), ",,"))
	}

	if strings.HasSuffix(*artifactName, ".war") {
		logrus.Infof(".war file detected. Using Apache Tomcat to deploy")
		argsStr += "-tomcat "
		tomcatCapstanFileContents := fmt.Sprintf(`
base: unik-tomcat

cmdline: /java.so %s -cp /usr/tomcat/bin/bootstrap.jar:usr/tomcat/bin/tomcat-juli.jar -jar /program.jar %s

#
# List of files that are included in the generated image.
#
files:
  /usr/tomcat/webapps/%s: %s`, proprtiesStr,
			argsStr,
			filepath.Base(artifactFile), artifactFile)
		if err := ioutil.WriteFile(filepath.Join(project_directory, "Capstanfile"), []byte(tomcatCapstanFileContents), 0644); err != nil {
			logrus.WithError(err).Error("failed writing capstanfile")
			os.Exit(-1)
		}
	} else if strings.HasSuffix(*artifactName, ".jar") {
		logrus.Infof("building Java unikernel from .jar file")
		argsStr += fmt.Sprintf("-jarName=/%s", filepath.Base(artifactFile))
		jarRunnerCapstanFileContents := fmt.Sprintf(`
base: unik-jar-runner

cmdline: /java.so %s -cp /%s -jar /program.jar %s

#
# List of files that are included in the generated image.
#
files:
  /%s: %s`, proprtiesStr,
			filepath.Base(artifactFile),
			argsStr,
			filepath.Base(artifactFile), artifactFile)
		if err := ioutil.WriteFile(filepath.Join(project_directory, "Capstanfile"), []byte(jarRunnerCapstanFileContents), 0644); err != nil {
			logrus.WithError(err).Error("failed writing capstanfile")
			os.Exit(-1)
		}
	} else {
		logrus.Errorf("%s is not of type .war or .jar, exiting!", *artifactName)
		os.Exit(-1)
	}

	go func() {
		fmt.Println("capstain building")

		capstanCmd := exec.Command("capstan", "run", "-p", "qemu")
		capstanCmd.Dir = project_directory
		capstanCmd.Stdout = os.Stdout
		capstanCmd.Stderr = os.Stderr
		printCommand(capstanCmd)
		if err := capstanCmd.Run(); err != nil {
			logrus.WithError(err).Error("capstan build failed")
			os.Exit(-1)
		}
	}()
	capstanImage := filepath.Join(os.Getenv("HOME"), ".capstan", "instances", "qemu", "project_directory", "disk.qcow2")

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
		project_directory+"/boot.qcow2")
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
