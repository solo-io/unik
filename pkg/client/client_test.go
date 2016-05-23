package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/Sirupsen/logrus"
)

var _ = Describe("Client", func() {
	var d *daemon.UnikDaemon
	BeforeEach(func(){
		var err error
		d, err = helpers.DaemonFromEnv()
		if err != nil {
			logrus.Fatalf(err)
		}
		go d.Run(3000)
	})
	AfterEach(func(){
		d.Stop()
		if err := helpers.KillUnikstate(); err != nil {
			logrus.Fatalf(err)
		}
	})
	Describe("images", func(){
		Describe("All()", func(){
			It("returns a list of images", func(){
				c := UnikClient("127.0.0.1:3000")
				images, err := c.Images().All()
				Expect(err).NotTo(HaveOccurred())
				Expect(images).To(BeEmpty())
			})
		})
	})
})
