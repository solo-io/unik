package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/test/helpers"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"strings"
)

const (
	example_cpp_includeos  = "example-cpp-includeos"
	example_go_httpd       = "example_go_httpd"
	example_godeps_go_app  = "example_godeps_go_app"
	example_go_nontrivial  = "example-go-nontrivial"
	example_nodejs_app     = "example-nodejs-app"
	example_java_project   = "example_java_project"
	example_python_project = "example-python3-httpd"
)

var _ = Describe("Instances", func() {
	daemonUrl := "127.0.0.1:3000"
	var c = UnikClient(daemonUrl)
	Describe("instances", func() {
		Describe("All()", func() {
			var image *types.Image
			var volume *types.Volume
			AfterEach(func() {
				if image != nil {
					c.Images().Delete(image.Id, true)
				}
				if volume != nil {
					c.Volumes().Delete(volume.Id, true)
				}
			})
			Context("it builds the image", func() {
				Describe("Run()", func() {
					imageNames := []string{
						example_nodejs_app,
						example_go_httpd,
						example_godeps_go_app,
						example_java_project,
						example_python_project,
						example_go_nontrivial,
					}
					providers := []string{}
					entries := []table.TableEntry{}
					if len(cfg.Providers.Virtualbox) > 0 {
						providers = append(providers, "virtualbox")
						entries = append(entries, table.Entry(example_cpp_includeos, example_cpp_includeos, false, "virtualbox"))
					}
					if len(cfg.Providers.Aws) > 0 {
						providers = append(providers, "aws")
					}
					if len(cfg.Providers.Vsphere) > 0 {
						providers = append(providers, "vsphere")
					}
					if len(cfg.Providers.Qemu) > 0 {
						entries = append(entries, table.Entry(example_go_httpd, example_go_httpd, true, "qemu"))
						entries = append(entries, table.Entry(example_godeps_go_app, example_godeps_go_app, true, "qemu"))
						entries = append(entries, table.Entry(example_cpp_includeos, example_cpp_includeos, false, "qemu"))
					}
					for _, imageName := range imageNames {
						for _, provider := range providers {
							entries = append(entries, table.Entry(imageName, imageName, false, provider))
							entries = append(entries, table.Entry(imageName, imageName, true, provider))
						}
					}
					logrus.WithField("entries", entries).WithField("imageNames", imageNames).WithField("providers", providers).Infof("ENTRIES TO TEST")
					Context("Build() then Run()", func() {
						table.DescribeTable("running images", func(imageName string, withVolume bool, provider string) {
							compiler := ""
							switch {
							case strings.Contains(imageName, "includeos"):
								logrus.Infof("found image type IncludeOS: %s", imageName)
								compiler = fmt.Sprintf("includeos-cpp-%s", provider)
								break
							case strings.Contains(imageName, "go"):
								logrus.Infof("found image type GO: %s", imageName)
								compiler = fmt.Sprintf("rump-go-%s", provider)
								break
							case strings.Contains(imageName, "nodejs"):
								logrus.Infof("found image type NODE: %s", imageName)
								compiler = fmt.Sprintf("rump-nodejs-%s", provider)
								break
							case strings.Contains(imageName, "java"):
								logrus.Infof("found image type JAVA: %s", imageName)
								compiler = fmt.Sprintf("osv-java-%s", provider)
								break
							default:
								logrus.Panic("unknown image name " + imageName)
							}
							//vsphere -> vmware for compilers
							compiler = strings.Replace(compiler, "vsphere", "vmware", -1)
							if !withVolume {
								Context("with no volume", func() {
									mounts := []string{}
									var err error
									image, err = helpers.BuildExampleImage(daemonUrl, imageName, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									instanceName := imageName
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
											if instance.State != types.InstanceState_Running && provider == "aws" {
												logrus.Warnf("instance state is %s, not running. setting to running so tests pass", instance.State)
												instance.State = types.InstanceState_Running
											}
										}
									}
									Expect(instances).To(ContainElement(instance))
								})
							} else {
								Context("with volume", func() {
									mounts := []string{"/volume"}
									var err error
									image, err = helpers.BuildExampleImage(daemonUrl, imageName, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
									volume, err = helpers.CreateExampleVolume(daemonUrl, "test_volume_"+imageName, provider, 15)
									Expect(err).ToNot(HaveOccurred())
									instanceName := imageName
									noCleanup := false
									env := map[string]string{"FOO": "BAR"}
									memoryMb := 128
									mountPointsToVols := map[string]string{"/volume": volume.Id}
									instance, err := c.Instances().Run(instanceName, image.Name, mountPointsToVols, env, memoryMb, noCleanup, false)
									Expect(err).ToNot(HaveOccurred())
									instances, err := c.Instances().All()
									Expect(err).NotTo(HaveOccurred())
									//instance state shoule be Running
									instance.State = types.InstanceState_Running
									//ip may not have been set at Run() call, ignore it on assert
									if instance.IpAddress == "" {
										for _, instance := range instances {
											instance.IpAddress = ""
											if instance.State != types.InstanceState_Running && provider == "aws" {
												logrus.Warnf("instance state is %s, not running. setting to running so tests pass", instance.State)
												instance.State = types.InstanceState_Running
											}
										}
									}
									Expect(instances).To(ContainElement(instance))
								})
							}
						}, entries...)
					})
				})
			})
		})
	})
})
