package out_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aditya87/precompiled-bosh-release-resource"
	"github.com/aditya87/precompiled-bosh-release-resource/compiler/fakes"
	"github.com/aditya87/precompiled-bosh-release-resource/out"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out Command", func() {
	var (
		boshClient        *fakes.BOSHClient
		manifestGenerator *fakes.ManifestGenerator
		command           *out.OutCommand
		boshTarget        string
		boshUser          string
		boshPassword      string
		releaseDirPath    string
		releaseVersion    string
		stemcellDirPath   string
		stemcellTarball   string
		request           out.OutRequest
	)

	BeforeEach(func() {
		boshTarget = "http://fake-bosh-target"
		boshUser = "fake-bosh-user"
		boshPassword = "fake-bosh-password"
		releaseVersion = "45"
		releaseDirPath, err := ioutil.TempDir("", "fake-bosh-release")
		Expect(err).ToNot(HaveOccurred())
		stemcellDirPath, err = ioutil.TempDir("", "stemcell-dir")
		Expect(err).ToNot(HaveOccurred())

		stemcellTarball = filepath.Join(stemcellDirPath, "some-stemcell-1.2.3.tgz")
		err = createStemcellTarball(stemcellTarball, bytes.NewBuffer([]byte(`---
operating_system: some-stemcell
version: 1.2.3
`)))
		Expect(err).NotTo(HaveOccurred())

		boshClient = &fakes.BOSHClient{}
		manifestGenerator = &fakes.ManifestGenerator{}

		request = out.OutRequest{
			Source: precompiled_release_resource.Source{
				BoshUser:     boshUser,
				BoshPassword: boshPassword,
				BoshTarget:   boshTarget,
			},
			Params: out.Params{
				ReleaseDir:     releaseDirPath,
				ReleaseVersion: releaseVersion,
				StemcellDir:    stemcellDirPath,
			},
		}

		command = out.NewOutCommand(request)
		command.BOSHClient = boshClient
	})

	AfterEach(func() {
		err := os.RemoveAll(stemcellDirPath)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(releaseDirPath)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("UploadStemcell", func() {
		It("uploads the stemcell to the bosh director", func() {
			err := command.UploadStemcell()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UploadStemcellCall.Receives.Contents).NotTo(BeNil())
			actualContents, err := ioutil.ReadAll(boshClient.UploadStemcellCall.Receives.Contents)
			Expect(err).NotTo(HaveOccurred())

			expectedContents, err := ioutil.ReadFile(stemcellTarball)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualContents).To(Equal(expectedContents))
		})
	})
})
