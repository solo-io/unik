package mirage

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMirage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mirage Suite")
}
