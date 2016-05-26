package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/Sirupsen/logrus"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	var projectRoot = helpers.GetProjectRoot()
	BeforeSuite(func(){
		logrus.SetLevel(logrus.DebugLevel)
		if err := helpers.MakeContainers(projectRoot); err != nil {
			logrus.Panic(err)
		}
	})
	AfterSuite(func(){
		//if err := helpers.RemoveContainers(projectRoot); err != nil {
		//	logrus.Panic(err)
		//}
	})
	RunSpecs(t, "Client Suite")
}
