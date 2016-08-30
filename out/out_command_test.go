package out_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aditya87/precompiled-bosh-release-resource"
	"github.com/aditya87/precompiled-bosh-release-resource/compiler/fakes"
	"github.com/aditya87/precompiled-bosh-release-resource/out"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"
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
		releaseName       string
	)

	BeforeEach(func() {
		var err error
		boshTarget = "http://fake-bosh-target"
		boshUser = "fake-bosh-user"
		boshPassword = "fake-bosh-password"
		releaseVersion = "45"
		releaseDirPath, err = ioutil.TempDir("", "fake-bosh-release")
		Expect(err).ToNot(HaveOccurred())
		err = os.Chmod(releaseDirPath, 0700)
		Expect(err).ToNot(HaveOccurred())

		err = os.Chdir(releaseDirPath)
		Expect(err).ToNot(HaveOccurred())

		err = os.Mkdir("blobs", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("config", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("jobs", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("packages", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("src", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("dev_releases", 0700)
		Expect(err).ToNot(HaveOccurred())
		err = os.Mkdir("dev_releases/foo", 0700)
		Expect(err).ToNot(HaveOccurred())

		stemcellDirPath, err = ioutil.TempDir("", "stemcell-dir")
		Expect(err).ToNot(HaveOccurred())

		stemcellTarball = filepath.Join(stemcellDirPath, "some-stemcell-1.2.3.tgz")
		err = createStemcellTarball(stemcellTarball, bytes.NewBuffer([]byte(`---
operating_system: some-stemcell
version: 1.2.3`)))
		Expect(err).NotTo(HaveOccurred())

		releaseTarballPath := filepath.Join(releaseDirPath, "dev_releases/foo/foo-42.tgz")
		err = createReleaseTarball(releaseTarballPath, bytes.NewBuffer([]byte(`---
name: foo
version: 42
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
		matches := regexp.MustCompile("(.*)/(.*)$").FindStringSubmatch(releaseDirPath)
		releaseName = matches[len(matches)-1]
	})

	AfterEach(func() {
		err := os.RemoveAll(stemcellDirPath)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(releaseDirPath)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("UploadStemcell", func() {
		Context("The stemcell does not already exist on the bosh director", func() {
			BeforeEach(func() {
				boshClient.StemcellCall.Returns.Stemcell = bosh.Stemcell{}
				boshClient.StemcellCall.Returns.Error = fmt.Errorf("stemcell some-stemcell could not be found")
			})

			It("uploads the stemcell to the bosh director", func() {
				err := command.UploadStemcell()
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClient.StemcellCall.CallCount).To(Equal(1))
				Expect(boshClient.StemcellCall.Receives).To(Equal("some-stemcell"))
				Expect(boshClient.UploadStemcellCall.Receives.Contents).NotTo(BeNil())
				actualContents, err := ioutil.ReadAll(boshClient.UploadStemcellCall.Receives.Contents)
				Expect(err).NotTo(HaveOccurred())

				expectedContents, err := ioutil.ReadFile(stemcellTarball)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualContents).To(Equal(expectedContents))
			})
		})

		Context("The stemcell already exists on the bosh director", func() {
			BeforeEach(func() {
				boshClient.StemcellCall.Returns.Stemcell = bosh.Stemcell{
					Name:     "some-stemcell",
					Versions: []string{"1.2.3"},
				}
				boshClient.StemcellCall.Returns.Error = nil
			})

			It("does not upload the stemcell to the bosh director", func() {
				err := command.UploadStemcell()
				Expect(err).NotTo(HaveOccurred())
				Expect(boshClient.StemcellCall.CallCount).To(Equal(1))
				Expect(boshClient.StemcellCall.Receives).To(Equal("some-stemcell"))
				Expect(boshClient.UploadStemcellCall.CallCount).To(Equal(0))
			})
		})
	})

	Describe("CreateRelease", func() {
		It("creates release with tarball", func() {
			err := command.CreateRelease()
			Expect(err).NotTo(HaveOccurred())
			expectedReleasePath := filepath.Join(releaseDirPath, fmt.Sprintf("dev_releases/%s/%s-%s.tgz", releaseName, releaseName, releaseVersion))
			Expect(expectedReleasePath).To(BeAnExistingFile())
			Expect(filepath.Join(releaseDirPath, "dev_releases/foo")).NotTo(BeADirectory())
		})
	})

	FDescribe("Run", func() {
		BeforeEach(func() {
			boshClient.InfoCall.Returns.DirectorInfo = bosh.DirectorInfo{
				UUID: "some-director-uuid",
			}
			manifestGenerator.GenerateCall.Returns.Manifest = []byte("deployment-manifest")
			boshClient.ExportReleaseCall.Returns.ResourceID = "some-resource-guid"
			boshClient.ResourceCall.Returns.Resource = ioutil.NopCloser(strings.NewReader("compiled-release-contents"))
		})

		It("deletes any pre-existing deployments", func() {
			boshClient.DeploymentsCall.Returns.DeploymentList = []bosh.Deployment{
				{Name: "dep1"},
				{Name: "dep2"},
			}
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(boshClient.DeploymentsCall.CallCount).To(Equal(1))
			Expect(len(boshClient.DeleteDeploymentCall.Receives.Name)).To(Equal(3))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[0]).To(Equal("dep1"))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[1]).To(Equal("dep2"))
		})

		It("uploads the release to the bosh director", func() {
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UploadReleaseCall.Receives.Contents).NotTo(BeNil())
			actualContents, err := ioutil.ReadAll(boshClient.UploadReleaseCall.Receives.Contents)
			Expect(err).NotTo(HaveOccurred())

			expectedContents, err := ioutil.ReadFile(filepath.Join(releaseDirPath, fmt.Sprintf("dev_releases/%s/%s-%s.tgz", releaseName, releaseName, releaseVersion)))
			Expect(err).NotTo(HaveOccurred())

			Expect(actualContents).To(Equal(expectedContents))
		})

		It("generates a deployment manifest", func() {
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestGenerator.GenerateCall.Receives.DirectorUUID).To(Equal("some-director-uuid"))
			Expect(manifestGenerator.GenerateCall.Receives.DeploymentName).To(Equal("compile-release-some-guid"))
			Expect(manifestGenerator.GenerateCall.Receives.Release.Name).To(Equal(releaseName))
			Expect(manifestGenerator.GenerateCall.Receives.Stemcell.Name).To(Equal("some-stemcell"))
		})

		It("deploys the manifest", func() {
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.DeployCall.Receives.Manifest).To(Equal([]byte("deployment-manifest")))
		})

		It("exports the release", func() {
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.ExportReleaseCall.Receives.DeploymentName).To(Equal("compile-release-some-guid"))
			Expect(boshClient.ExportReleaseCall.Receives.ReleaseName).To(Equal(releaseName))
			Expect(boshClient.ExportReleaseCall.Receives.ReleaseVersion).To(Equal(releaseVersion))
			Expect(boshClient.ExportReleaseCall.Receives.StemcellName).To(Equal("some-stemcell"))
			Expect(boshClient.ExportReleaseCall.Receives.StemcellVersion).To(Equal("1.2.3"))
		})

		It("deletes the deployment", func() {
			err := command.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(boshClient.DeleteDeploymentCall.Receives.Name)).To(Equal(1))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[0]).To(Equal("compile-release-some-guid"))
		})
	})

})
