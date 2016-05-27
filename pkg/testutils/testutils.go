package testutils

import (
	"io/ioutil"
	"os"

	"github.com/emc-advanced-dev/unik/pkg/config"
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
