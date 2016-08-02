package out_test

import (
	"io/ioutil"

	"github.com/aditya87/precompiled-bosh-release-resource"
	"github.com/aditya87/precompiled-bosh-release-resource/out"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out Command", func() {
	Describe("Happy path", func() {
		var (
			boshTarget      string
			boshUser        string
			boshPassword    string
			releaseDirPath  string
			stemcellName    string
			stemcellVersion string
			request         out.OutRequest
		)

		BeforeEach(func() {
			var err error
			releaseDirPath, err = ioutil.TempDir("", "fake-bosh-release")
			Expect(err).ToNot(HaveOccurred())

			boshTarget = "fake-bosh-target"
			boshUser = "fake-bosh-user"
			boshPassword = "fake-bosh-password"
			stemcellName = "fake-stemcell"
			stemcellVersion = "fake-stemcell-version"

			request = out.OutRequest{
				Source: precompiled_release_resource.Source{
					BoshUser:     boshUser,
					BoshPassword: boshPassword,
					BoshTarget:   boshTarget,
				},
				Params: out.Params{
					ReleaseDir:      releaseDirPath,
					StemcellName:    stemcellName,
					StemcellVersion: stemcellVersion,
				},
			}

		})
	})
})
