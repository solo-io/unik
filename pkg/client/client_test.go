package client_test

import (
	. "github.com/emc-advanced-dev/unik/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	It("runs the test", func(){
		c := UnikClient("127.0.0.1")
		images, err := c.Images().All()
		Expect(err).To(HaveOccurred())
		Expect(images).To(BeNil())
	})
})
