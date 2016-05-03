package main

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"io/ioutil"
	"strings"
	"github.com/Sirupsen/logrus"
	"encoding/xml"
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
)

type AppInfo struct {
	GroupId string
	ArtifactId string
	Version string
}

func wrapJavaApplication(javaWrapperDir, appSourceDir string) (AppInfo, error) {
	logrus.Info("start1")
	appPom, err := readPom(appSourceDir + "/pom.xml")
	if err != nil {
		return AppInfo{}, errors.New("reading pom", err)
	}
	logrus.Info("start2")

	groupId := appPom.GroupId.Text
	artifactId := appPom.ArtifactId.Text
	version := appPom.Version.Text

	wrapperPomBytes, err := ioutil.ReadFile(javaWrapperDir + "/pom.xml")
	if err != nil {
		return AppInfo{}, errors.New("reading app pom bytes", err)
	}
	wrapperPomContents := strings.Replace(string(wrapperPomBytes), "REPLACE_WITH_GROUPID", groupId, -1)
	wrapperPomContents = strings.Replace(wrapperPomContents, "REPLACE_WITH_ARTIFACTID", artifactId, -1)
	wrapperPomContents = strings.Replace(wrapperPomContents, "REPLACE_WITH_VERSION", version, -1)

	err = ioutil.WriteFile(javaWrapperDir + "/pom.xml", []byte(wrapperPomContents), 0666)
	if err != nil {
		return AppInfo{}, errors.New("writing pom.xml", err)
	}

	mainClassName, err := appPom.getMainClass()
	if err != nil {
		return AppInfo{}, errors.New("retreiving main class from app", err)
	}
	logrus.WithFields(logrus.Fields{
		"pom": appPom,
		"groupid": appPom.GroupId,
		"artifactId": appPom.ArtifactId,
		"version": appPom.Version,
		"mainClassName": mainClassName,
	}).Infof("parsed app pom.xml, gathered relevant fields")

	wrapperMainContentBytes, err := ioutil.ReadFile(javaWrapperDir + "/src/main/java/com/emc/wrapper/Wrapper.java")
	if err != nil {
		return AppInfo{}, errors.New("reading java pom bytes", err)
	}
	wrapperMainContents := strings.Replace(string(wrapperMainContentBytes), "REPLACE_WITH_MAIN_CLASS", mainClassName, -1)

	err = ioutil.WriteFile(javaWrapperDir + "/src/main/java/com/emc/wrapper/Wrapper.java", []byte(wrapperMainContents), 0666)
	if err != nil {
		return AppInfo{}, errors.New("writing Wrapper class around app class", err)
	}

	return AppInfo{ArtifactId: artifactId, GroupId: groupId, Version: version}, nil
}


func readPom(filename string) (*project, error) {
	reader, xmlFile, err := genericReader(filename)
	if err != nil {
		return nil, err
	}
	if xmlFile != nil {
		defer xmlFile.Close()
	}

	decoder := xml.NewDecoder(reader)
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch se := token.(type) {
		case xml.StartElement:
			return handleFeed(se, decoder)
		}
	}
	return nil, errors.New("decoding failed", nil)
}


func handleFeed(se xml.StartElement, decoder *xml.Decoder) (*project, error) {
	if se.Name.Local == "project" {
		var item project
		decoder.DecodeElement(&item, &se)
		return &item, nil
	}
	return nil, errors.New(se.Name.Local+"not a project", nil)
}

func genericReader(filename string) (io.Reader, *os.File, error) {
	if filename == "" {
		return bufio.NewReader(os.Stdin), nil, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	if strings.HasSuffix(filename, "bz2") {
		return bufio.NewReader(bzip2.NewReader(bufio.NewReader(file))), file, err
	}

	if strings.HasSuffix(filename, "gz") {
		reader, err := gzip.NewReader(bufio.NewReader(file))
		if err != nil {
			return nil, nil, err
		}
		return bufio.NewReader(reader), file, err
	}
	return bufio.NewReader(file), file, err
}

type root struct {
	Project *project `xml:"http://maven.apache.org/POM/4.0.0 project,omitempty" json:"project,omitempty"`
}

type project struct {
	Attr_xmlns              string `xml:" xmlns,attr"  json:",omitempty"`
	Attr_xsi                string `xml:"xmlns xsi,attr"  json:",omitempty"`
	Attr_xsi_schemaLocation string `xml:"http://www.w3.org/2001/XMLSchema-instance schemaLocation,attr"  json:",omitempty"`
	ArtifactId              *ArtifactId `xml:"http://maven.apache.org/POM/4.0.0 artifactId,omitempty" json:"artifactId,omitempty"`
	Build        *build `xml:"http://maven.apache.org/POM/4.0.0 build,omitempty" json:"build,omitempty"`
	Dependencies *dependencies `xml:"http://maven.apache.org/POM/4.0.0 dependencies,omitempty" json:"dependencies,omitempty"`
	GroupId      *GroupId `xml:"http://maven.apache.org/POM/4.0.0 groupId,omitempty" json:"groupId,omitempty"`
	ModelVersion *modelVersion `xml:"http://maven.apache.org/POM/4.0.0 modelVersion,omitempty" json:"modelVersion,omitempty"`
	Name         *name `xml:"http://maven.apache.org/POM/4.0.0 name,omitempty" json:"name,omitempty"`
	Packaging    *packaging `xml:"http://maven.apache.org/POM/4.0.0 packaging,omitempty" json:"packaging,omitempty"`
	Properties   *properties `xml:"http://maven.apache.org/POM/4.0.0 properties,omitempty" json:"properties,omitempty"`
	Url          *url `xml:"http://maven.apache.org/POM/4.0.0 url,omitempty" json:"url,omitempty"`
	Version      *Version `xml:"http://maven.apache.org/POM/4.0.0 version,omitempty" json:"version,omitempty"`
	XMLName      xml.Name `xml:"http://maven.apache.org/POM/4.0.0 project,omitempty" json:"project,omitempty"`
}

type build struct {
	Plugins *plugins `xml:"http://maven.apache.org/POM/4.0.0 plugins,omitempty" json:"plugins,omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 build,omitempty" json:"build,omitempty"`
}

type plugins struct {
	Plugin  []*plugin `xml:"http://maven.apache.org/POM/4.0.0 plugin,omitempty" json:"plugin,omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 plugins,omitempty" json:"plugins,omitempty"`
}

type plugin struct {
	ArtifactId    *ArtifactId `xml:"http://maven.apache.org/POM/4.0.0 artifactId,omitempty" json:"artifactId,omitempty"`
	Configuration *configuration `xml:"http://maven.apache.org/POM/4.0.0 configuration,omitempty" json:"configuration,omitempty"`
	Executions    *executions `xml:"http://maven.apache.org/POM/4.0.0 executions,omitempty" json:"executions,omitempty"`
	GroupId       *GroupId `xml:"http://maven.apache.org/POM/4.0.0 groupId,omitempty" json:"groupId,omitempty"`
	Version       *Version `xml:"http://maven.apache.org/POM/4.0.0 version,omitempty" json:"version,omitempty"`
	XMLName       xml.Name `xml:"http://maven.apache.org/POM/4.0.0 plugin,omitempty" json:"plugin,omitempty"`
}

type executions struct {
	Execution *execution `xml:"http://maven.apache.org/POM/4.0.0 execution,omitempty" json:"execution,omitempty"`
	XMLName   xml.Name `xml:"http://maven.apache.org/POM/4.0.0 executions,omitempty" json:"executions,omitempty"`
}

type execution struct {
	Goals   *goals `xml:"http://maven.apache.org/POM/4.0.0 goals,omitempty" json:"goals,omitempty"`
	Phase   *phase `xml:"http://maven.apache.org/POM/4.0.0 phase,omitempty" json:"phase,omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 execution,omitempty" json:"execution,omitempty"`
}

type phase struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 phase,omitempty" json:"phase,omitempty"`
}

type goals struct {
	Goal    *goal `xml:"http://maven.apache.org/POM/4.0.0 goal,omitempty" json:"goal,omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 goals,omitempty" json:"goals,omitempty"`
}

type goal struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 goal,omitempty" json:"goal,omitempty"`
}

type GroupId struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 groupId,omitempty" json:"groupId,omitempty"`
}

type ArtifactId struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 artifactId,omitempty" json:"artifactId,omitempty"`
}

type Version struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 version,omitempty" json:"version,omitempty"`
}

type configuration struct {
	Archive        *archive `xml:"http://maven.apache.org/POM/4.0.0 archive,omitempty" json:"archive,omitempty"`
	DescriptorRefs *descriptorRefs `xml:"http://maven.apache.org/POM/4.0.0 descriptorRefs,omitempty" json:"descriptorRefs,omitempty"`
	XMLName        xml.Name `xml:"http://maven.apache.org/POM/4.0.0 configuration,omitempty" json:"configuration,omitempty"`
}

type descriptorRefs struct {
	DescriptorRef *descriptorRef `xml:"http://maven.apache.org/POM/4.0.0 descriptorRef,omitempty" json:"descriptorRef,omitempty"`
	XMLName       xml.Name `xml:"http://maven.apache.org/POM/4.0.0 descriptorRefs,omitempty" json:"descriptorRefs,omitempty"`
}

type descriptorRef struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 descriptorRef,omitempty" json:"descriptorRef,omitempty"`
}

type archive struct {
	Manifest *manifest `xml:"http://maven.apache.org/POM/4.0.0 manifest,omitempty" json:"manifest,omitempty"`
	XMLName  xml.Name `xml:"http://maven.apache.org/POM/4.0.0 archive,omitempty" json:"archive,omitempty"`
}

type manifest struct {
	MainClass *mainClass `xml:"http://maven.apache.org/POM/4.0.0 mainClass,omitempty" json:"mainClass,omitempty"`
	XMLName   xml.Name `xml:"http://maven.apache.org/POM/4.0.0 manifest,omitempty" json:"manifest,omitempty"`
}

type mainClass struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 mainClass,omitempty" json:"mainClass,omitempty"`
}

type modelVersion struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 modelVersion,omitempty" json:"modelVersion,omitempty"`
}

type packaging struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 packaging,omitempty" json:"packaging,omitempty"`
}

type url struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 url,omitempty" json:"url,omitempty"`
}

type dependencies struct {
	Dependency []*dependency `xml:"http://maven.apache.org/POM/4.0.0 dependency,omitempty" json:"dependency,omitempty"`
	XMLName    xml.Name `xml:"http://maven.apache.org/POM/4.0.0 dependencies,omitempty" json:"dependencies,omitempty"`
}

type dependency struct {
	ArtifactId *ArtifactId `xml:"http://maven.apache.org/POM/4.0.0 artifactId,omitempty" json:"artifactId,omitempty"`
	GroupId    *GroupId `xml:"http://maven.apache.org/POM/4.0.0 groupId,omitempty" json:"groupId,omitempty"`
	Scope      *scope `xml:"http://maven.apache.org/POM/4.0.0 scope,omitempty" json:"scope,omitempty"`
	Version    *Version `xml:"http://maven.apache.org/POM/4.0.0 version,omitempty" json:"version,omitempty"`
	XMLName    xml.Name `xml:"http://maven.apache.org/POM/4.0.0 dependency,omitempty" json:"dependency,omitempty"`
}

type scope struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 scope,omitempty" json:"scope,omitempty"`
}

type name struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 name,omitempty" json:"name,omitempty"`
}

type properties struct {
	Project_dot_build_dot_sourceEncoding *project_dot_build_dot_sourceEncoding `xml:"http://maven.apache.org/POM/4.0.0 project.build.sourceEncoding,omitempty" json:"project.build.sourceEncoding,omitempty"`
	XMLName                              xml.Name `xml:"http://maven.apache.org/POM/4.0.0 properties,omitempty" json:"properties,omitempty"`
}

type project_dot_build_dot_sourceEncoding struct {
	Text    string `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://maven.apache.org/POM/4.0.0 project.build.sourceEncoding,omitempty" json:"project.build.sourceEncoding,omitempty"`
}



func (project *project) getMainClass() (string, error) {
	if project.Build != nil {
		for _, plugin := range project.Build.Plugins.Plugin {
			if plugin.Configuration != nil &&
			plugin.Configuration != nil &&
			plugin.Configuration.Archive != nil &&
			plugin.Configuration.Archive.Manifest != nil &&
			plugin.Configuration.Archive.Manifest.MainClass != nil {
				return plugin.Configuration.Archive.Manifest.MainClass.Text, nil
			}
		}
	}
	return "", errors.New("main class not found", nil)
}