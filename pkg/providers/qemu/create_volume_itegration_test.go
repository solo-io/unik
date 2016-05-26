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

type TempUnikHome struct {
	Dir string
}

func (t *TempUnikHome) setupUnik() {
	n, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	config.Internal.UnikHome = n

	t.Dir = n
}

func (t *TempUnikHome) tearDownUnik() {
	os.RemoveAll(t.Dir)
}

var _ = Describe("QemuProvider", func() {

	var tmpUnik TempUnikHome

	BeforeEach(func() {
		tmpUnik.setupUnik()
	})

	AfterEach(func() {
		tmpUnik.tearDownUnik()
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
