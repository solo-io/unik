// +build integration

package qemu_test

import (
	"io/ioutil"
	"os"

	"github.com/emc-advanced-dev/unik/pkg/config"
	. "github.com/emc-advanced-dev/unik/pkg/providers/qemu"
	"github.com/emc-advanced-dev/unik/pkg/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("QemuProvider", func() {
	var tmpUnik helpers.TempUnikHome

	BeforeEach(func() {
		tmpUnik.SetupUnik()
	})

	AfterEach(func() {
		tmpUnik.TearDownUnik()
	})

	It("should create a volume", func() {
		config := config.Qemu{Name: "test"}
		q, err := NewQemuProvider(config)
		Expect(err).NotTo(HaveOccurred())

		f, err := ioutil.TempFile(tmpUnik.Dir, "")
		Expect(err).NotTo(HaveOccurred())

		_, err = f.WriteAt([]byte{'y'}, 1024)
		Expect(err).NotTo(HaveOccurred())

		f.Close()

		params := types.CreateVolumeParams{
			Name:      "testvol",
			ImagePath: f.Name(),
		}

		v, err := q.CreateVolume(params)

		Expect(err).NotTo(HaveOccurred())

		Expect(v.Id).ToNot(Equal(""))

	})
})
