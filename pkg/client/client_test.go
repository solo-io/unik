package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/Sirupsen/logrus"
	"os"
)

var _ = Describe("Client", func() {
	var d *daemon.UnikDaemon
	BeforeSuite(func(){
		Describe("building containers", func(){
			It("builds all compilers and utils in containers", func(){
				projectRoot := os.Getenv("PROJECT_ROOT")
				if projectRoot == "" {
					var err error
					projectRoot, err = os.Getwd() //requires running ginkgo from project root already
					Expect(err).NotTo(HaveOccurred())
				}
				err := helpers.MakeContainers(projectRoot)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	AfterSuite(func(){
		Describe("removing containers", func(){
			It("removes all compiler and util containers", func(){
				projectRoot := os.Getenv("PROJECT_ROOT")
				if projectRoot == "" {
					var err error
					projectRoot, err = os.Getwd() //requires running ginkgo from project root already
					Expect(err).NotTo(HaveOccurred())
				}
				err := helpers.RemoveContainers(projectRoot)
				Expect(err).NotTo(HaveOccurred())
			})

		})
	})
	BeforeEach(func(){
		Describe("start the daeemon", func(){
			It("deploys the instance listener and starts listening on port 3000", func(){
				var err error
				d, err = helpers.DaemonFromEnv()
				Expect(err).ToNot(HaveOccurred())
				go d.Run(3000)
			})

		})
	})
	AfterEach(func(){
		It("tears down the unik daemon and cleans up the state", func(){
			err := d.Stop()
			Expect(err).ToNot(HaveOccurred())
			if err := helpers.KillUnikstate(); err != nil {
				logrus.Fatalf(err)
			}
		})
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
