package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"net/http"
	"os"
	"testing"
	"time"
)

var cfg = helpers.NewTestConfig()

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	var d *daemon.UnikDaemon
	var tmpUnik helpers.TempUnikHome
	BeforeSuite(func() {
		if os.Getenv("DEBUG_OFF") != "1" {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if os.Getenv("MAKE_CONTAINERS") == "1" {
			if err := helpers.MakeContainers(helpers.GetProjectRoot()); err != nil {
				logrus.Panic(err)
			}
		}

		tmpUnik.SetupUnik()
		var err error
		d, err = daemon.NewUnikDaemon(cfg)
		if err != nil {
			logrus.Panic(err)
		}
		go d.Run(3000)

	})
	AfterSuite(func() {
		//if err := helpers.RemoveContainers(projectRoot); err != nil {
		//	logrus.Panic(err)
		//}
		defer tmpUnik.TearDownUnik()
		err := d.Stop()
		if err != nil {
			logrus.Panic(err)
		}
	})
	RunSpecs(t, "Client Suite")
}

func testInstancePing(instanceIp string) {
	testInstanceEndpoint(instanceIp, "/ping_test", "pong")
}

func testInstanceEnv(instanceIp string) {
	testInstanceEndpoint(instanceIp, "/env_test", "VAL")
}

func testInstanceMount(instanceIp string) {
	testInstanceEndpoint(instanceIp, "/mount_test", "test_data")
}

func testInstanceEndpoint(instanceIp, path, expectedResponse string) {
	var resp *http.Response
	var body []byte
	var err error
	err = util.Retry(10, 2*time.Second, func() error {
		resp, body, err = lxhttpclient.Get(instanceIp+":8080", path, nil)
		return err
	})
	logrus.WithFields(logrus.Fields{
		"resp": resp,
		"body": string(body),
		"err":  err,
	}).Debugf("got resp")
	Expect(err).ToNot(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	var testResponse struct {
		Message string `json:"message"`
	}
	err = json.Unmarshal(body, &testResponse)
	Expect(err).ToNot(HaveOccurred())
	Expect(testResponse.Message).To(ContainSubstring(expectedResponse))
}
