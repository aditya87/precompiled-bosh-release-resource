package out_test

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out Command", func() {
	Describe("Happy path", func() {
		var (
			boshTarget     string
			boshUser       string
			boshPassword   string
			releaseDirPath string
		)

		BeforeEach(func() {
			var err error
			releaseDirPath, err = ioutil.TempDir("", "out_command_release")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
