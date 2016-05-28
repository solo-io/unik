package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	var projectRoot = helpers.GetProjectRoot()
	var d *daemon.UnikDaemon
	var tmpUnik helpers.TempUnikHome
	var cfg = helpers.NewTestConfig()
	BeforeSuite(func(){
		logrus.SetLevel(logrus.DebugLevel)
		if err := helpers.MakeContainers(projectRoot); err != nil {
			logrus.Panic(err)
		}
		util.SetContainerVer("1.0")

		tmpUnik.SetupUnik()
		var err error
		d, err = daemon.NewUnikDaemon(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
		go d.Run(3000)

	})
	AfterSuite(func(){
		//if err := helpers.RemoveContainers(projectRoot); err != nil {
		//	logrus.Panic(err)
		//}
		defer tmpUnik.TearDownUnik()
		err := d.Stop()
		if err != nil {
			logrus.Fatal(err)
		}
	})
	RunSpecs(t, "Client Suite")
}
