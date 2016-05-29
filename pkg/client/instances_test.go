package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

var _ = Describe("Instances", func() {
	daemonUrl := "127.0.0.1:3000"
	var c = UnikClient(daemonUrl)
	var projectRoot = helpers.GetProjectRoot()
	Describe("instances", func() {
		Describe("All()", func() {
			var image *types.Image
			var volume *types.Volume
			AfterEach(func() {
				if image != nil {
					if err := c.Images().Delete(image.Id, true); err != nil {
						logrus.Panic(err)
					}
				}
				if volume != nil {
					if err := c.Volumes().Delete(volume.Id, true); err != nil {
						logrus.Panic(err)
					}
				}
			})
			Context("no instances exist", func() {
				if len(cfg.Providers.Virtualbox) > 0 && len(cfg.Providers.Vsphere) < 1 ||
				   len(cfg.Providers.Virtualbox) < 1 && len(cfg.Providers.Vsphere) > 0 {
					Context("on virtualbox or vsphere provider", func(){
						It("returns a list with only the Instance Listener VM", func() {
							instances, err := c.Instances().All()
							Expect(err).NotTo(HaveOccurred())
							Expect(instances).To(HaveLen(1))
							Expect(instances[0].Name).To(ContainSubstring("Listener"))
						})

					})
				} else if len(cfg.Providers.Virtualbox) > 0 && len(cfg.Providers.Vsphere) > 0 {
					Context("on virtualbox and vsphere providers", func(){
						It("returns a list with only the Instance Listener VMs", func() {
							instances, err := c.Instances().All()
							Expect(err).NotTo(HaveOccurred())
							Expect(instances).To(HaveLen(2))
							Expect(instances[0].Name).To(ContainSubstring("Listener"))
							Expect(instances[1].Name).To(ContainSubstring("Listener"))
						})

					})
				} else if len(cfg.Providers.Virtualbox) < 1 && len(cfg.Providers.Vsphere) < 1 {
					It("returns an empty list", func() {
						instances, err := c.Instances().All()
						Expect(err).NotTo(HaveOccurred())
						Expect(instances).To(BeEmpty())
					})
				}
			})
			Context("instances exist", func(){
				Describe("Run()", func(){
					Context("with virtualbox as provider", func(){
						provider := "virtualbox"
						Context("with go app", func(){
							compiler := "rump-go-virtualbox"
							Context("with no volume", func(){
								It("runs successfully", func(){
									mounts := []string{}
									var err error
									image, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									instanceName := example_go_httpd
									volsToMounts := map[string]string{}
									instance, err := helpers.RunExampleInstance(daemonUrl, instanceName, image.Name, volsToMounts)
									Expect(err).ToNot(HaveOccurred())
									instances, err := c.Instances().All()
									Expect(err).NotTo(HaveOccurred())
									//instance state shoule be Running
									instance.State = types.InstanceState_Running
									//ip may not have been set at Run() call, ignore it on assert
									if instance.IpAddress == "" {
										for _, instance := range instances {
											instance.IpAddress = ""
										}
									}
									Expect(instances).To(ContainElement(instance))
								})
							})
							Context("with volume", func(){
								It("runs successfully and mounts the volume", func(){
									mounts := []string{"/volume"}
									var err error
									image, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									volume, err = helpers.CreateExampleVolume(daemonUrl, "test_volume", provider, 15)
									Expect(err).ToNot(HaveOccurred())
									instanceName := example_go_httpd
									noCleanup := false
									env := map[string]string{"FOO": "BAR"}
									memoryMb := 128
									mountPointsToVols := map[string]string{ "/volume": volume.Id}
									instance, err := c.Instances().Run(instanceName, image.Name, mountPointsToVols, env, memoryMb, noCleanup)
									Expect(err).ToNot(HaveOccurred())
									instances, err := c.Instances().All()
									Expect(err).NotTo(HaveOccurred())
									//instance state shoule be Running
									instance.State = types.InstanceState_Running
									//ip may not have been set at Run() call, ignore it on assert
									if instance.IpAddress == "" {
										for _, instance := range instances {
											instance.IpAddress = ""
										}
									}
									Expect(instances).To(ContainElement(instance))
								})
							})
						})
					})
				})
			})
		})
	})
})
