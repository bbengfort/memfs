package memfs_test

import (
	. "github.com/bbengfort/memfs"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MemFS Package", func() {

	const ExpectedVersion = "0.2.2"

	It("should have the right version", func() {
		Ω(PackageVersion()).Should(Equal(ExpectedVersion))
	})

})
