package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/Sirupsen/logrus"
)

var _ = Describe("Instances", func() {
	daemonUrl := "127.0.0.1:3000"
	var c = UnikClient(daemonUrl)
	var projectRoot = helpers.GetProjectRoot()
	FDescribe("instances", func() {
		Describe("All()", func() {
			AfterEach(func(){
				images, err := c.Images().All()
				if err != nil {
					logrus.Panic(err)
				}
				for _, image := range images {
					err = c.Images().Delete(image.Id, true)
					if err != nil {
						logrus.Panic(err)
					}
				}
				instances, err := c.Instances().All()
				if err != nil {
					logrus.Panic(err)
				}
				for _, instance := range instances {
					err = c.Instances().Stop(instance.Id)
					if err != nil {
						logrus.Panic(err)
					}
				}
				volumes, err := c.Volumes().All()
				if err != nil {
					logrus.Panic(err)
				}
				for _, volume := range volumes {
					err = c.Volumes().Delete(volume.Id, true)
					if err != nil {
						logrus.Panic(err)
					}
				}
			})
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
						provider := "virtualbox"
						Context("with go app", func(){
							compiler := "rump-go-virtualbox"
							Context("with no volume", func(){
								It("runs successfully", func(){
									mounts := []string{}
									image, err := helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									instanceName := example_go_httpd
									volsToMounts := map[string]string{}
									instance, err := helpers.RunExampleInstance(daemonUrl, instanceName, image.Name, volsToMounts)
									Expect(err).ToNot(HaveOccurred())
									instances, err := c.Instances().All()
									Expect(err).NotTo(HaveOccurred())
									Expect(instances).To(ContainElement(instance))
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
									mountPointsToVols := map[string]string{ "/volume": volume.Id}
									instance, err := c.Instances().Run(instanceName, image.Name, mountPointsToVols, env, memoryMb, noCleanup)
									Expect(err).ToNot(HaveOccurred())
									instances, err := c.Instances().All()
									Expect(err).NotTo(HaveOccurred())
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
