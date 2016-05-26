package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/emc-advanced-dev/unik/test/helpers"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

const (
	example_go_httpd = "example_go_httpd"
	example_godeps_go_app = "example_godeps_go_app"
	example_go_nontrivial = "example_go_nontrivial"
	example_nodejs_app = "example_nodejs_app"
	example_java_project = "example_java_project"
)

var _ = Describe("Images", func() {
	var d *daemon.UnikDaemon
	var daemonUrl = "127.0.0.1:3000"
	var c = UnikClient(daemonUrl)
	var projectRoot = os.Getenv("PROJECT_ROOT")
	var tmpUnik helpers.TempUnikHome
	var cfg = helpers.NewTestConfig()

	BeforeEach(func(){
		Describe("start the daeemon", func(){
			tmpUnik.SetupUnik()
			var err error
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
	Describe("images", func(){
		Describe("All()", func(){
			Context("no images exist", func(){
				It("returns an empty list", func(){
					images, err := c.Images().All()
					Expect(err).NotTo(HaveOccurred())
					Expect(images).To(BeEmpty())
				})
			})
			Context("images exist", func(){
				var simpleGoImage, dependencyGoImage, nontrivialGoImage, nodejsImage, javaImage *types.Image
				Describe("Build", func(){
					provider := "virtualbox"
					Context("using virtualbox provider", func(){
						Context("go apps", func(){
							compiler := "rump-go-virtualbox"
							Context("a simple go httpd with no dependencies", func(){
								mounts := []string{}
								It("compiles the app", func(){
									var err error
									simpleGoImage, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_httpd, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
							Context("a go app with dependencies", func(){
								mounts := []string{}
								It("compiles the app", func(){
									var err error
									dependencyGoImage, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_godeps_go_app, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
							Context("a go app with nested packages", func(){
								mounts := []string{}
								It("compiles the app", func(){
									var err error
									nontrivialGoImage, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_go_nontrivial, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
						})
						Context("node apps", func(){
							compiler := "rump-nodejs-virtualbox"
							Context("a node app with dependencies", func(){
								mounts := []string{}
								It("compiles the app", func(){
									var err error
									nodejsImage, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_nodejs_app, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
						})
						Context("java apps", func(){
							compiler := "osv-java-virtualbox"
							Context("a java app with dependencies", func(){
								mounts := []string{}
								It("compiles the app", func(){
									var err error
									javaImage, err = helpers.BuildExampleImage(daemonUrl, projectRoot, example_java_project, compiler, provider, mounts)
									Expect(err).ToNot(HaveOccurred())
								})
							})
						})
					})
				})
				It("returns the image as part of a list", func(){
					images, err := c.Images().All()
					Expect(err).NotTo(HaveOccurred())
					Expect(images).To(ContainElement(simpleGoImage))
					Expect(images).To(ContainElement(dependencyGoImage))
					Expect(images).To(ContainElement(nontrivialGoImage))
					Expect(images).To(ContainElement(nodejsImage))
					Expect(images).To(ContainElement(javaImage))
				})
				Describe("Delete()", func(){
					It("completely removes the image", func(){
						var err error
						err = c.Images().Delete(example_go_httpd, true)
						Expect(err).ToNot(HaveOccurred())
						c.Images().Delete(example_go_nontrivial, true)
						Expect(err).ToNot(HaveOccurred())
						c.Images().Delete(example_godeps_go_app, true)
						Expect(err).ToNot(HaveOccurred())
						c.Images().Delete(example_java_project, true)
						Expect(err).ToNot(HaveOccurred())
						c.Images().Delete(example_nodejs_app, true)
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})
		})
	})
})
