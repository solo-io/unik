package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
)

var _ = Describe("Instances", func() {
	var d *daemon.UnikDaemon
	daemonUrl := "127.0.0.1:3000"
	var c = UnikClient(daemonUrl)
	var projectRoot = helpers.GetProjectRoot()
	var tmpUnik helpers.TempUnikHome

	BeforeEach(func(){
		Describe("start the daeemon", func(){
			tmpUnik.SetupUnik()
			vboxConfig, err := helpers.NewVirtualboxConfig()
			if err != nil {
				logrus.Fatal(err)
			}
			cfg := helpers.ConfigWithVirtualbox(config.DaemonConfig{}, vboxConfig)
			d, err = daemon.NewUnikDaemon(cfg)
			if err != nil {
				logrus.Fatal(err)
			}
			go d.Run(3000)
		})
	})
	AfterEach(func() {
		defer tmpUnik.TearDownUnik()
		err := d.Stop()
		if err != nil {
			logrus.Fatal(err)
		}
	})
	Describe("instances", func() {
		Describe("All()", func() {
			AfterEach(func(){
				It("cleans up all images", func(){
					images, err := c.Images().All()
					Expect(err).NotTo(HaveOccurred())
					for _, image := range images {
						err = c.Images().Delete(image.Id, true)
						Expect(err).NotTo(HaveOccurred())
					}
				})
				It("cleans up all volumes", func(){
					instances, err := c.Instances().All()
					Expect(err).NotTo(HaveOccurred())
					for _, instance := range instances {
						err = c.Instances().Stop(instance.Id)
						Expect(err).NotTo(HaveOccurred())
					}
					volumes, err := c.Volumes().All()
					Expect(err).NotTo(HaveOccurred())
					for _, volume := range volumes {
						err = c.Volumes().Delete(volume.Id, true)
						Expect(err).NotTo(HaveOccurred())
					}
				})
			})
			var instanceGoNoVolume, instanceGoWithVolume *types.Instance
			Context("no instances exist", func() {
				It("returns an empty list", func() {
					instances, err := c.Instances().All()
					Expect(err).NotTo(HaveOccurred())
					Expect(instances).To(BeEmpty())
				})
			})
			Context("instances exist", func(){
				Describe("Run()", func(){
					Context("with virtualbox as provider", func(){
						provider := "provider"
						Context("with go app", func(){
							compiler := "rump-go-virtualbox"
							Context("with no volume", func(){
								It("runs successfully", func(){
									mounts := []string{}
									image, err := helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									instanceName := example_go_httpd
									volsToMounts := map[string]string{}
									instanceGoNoVolume, err = helpers.RunExampleInstance(daemonUrl, instanceName, image.Name, volsToMounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
							Context("with volume", func(){
								It("runs successfully and mounts the volume", func(){
									mounts := []string{"/volume"}
									image, err := helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									volume, err := helpers.CreateExampleVolume(daemonUrl, "test_volume", provider, 15)
									Expect(err).ToNot(HaveOccurred())
									instanceName := example_go_httpd
									noCleanup := false
									env := map[string]string{"FOO": "BAR"}
									memoryMb := 128
									volsToMounts := map[string]string{volume.Id: "/volume"}
									instanceGoNoVolume, err = c.Instances().Run(instanceName, image.Name, volsToMounts, env, memoryMb, noCleanup)
									Expect(err).ToNot(HaveOccurred())
								})
							})
						})
					})
				})
				It("lists all instances", func(){
					instances, err := c.Instances().All()
					Expect(err).NotTo(HaveOccurred())
					Expect(instances).To(ContainElement(instanceGoNoVolume))
					Expect(instances).To(ContainElement(instanceGoWithVolume))
				})
			})
		})
	})
})
