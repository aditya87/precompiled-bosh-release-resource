package compiler_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
	"github.com/pivotal-cf/pcf-releng-ci/tasks/future/compile-release/compiler"
	"github.com/pivotal-cf/pcf-releng-ci/tasks/future/compile-release/compiler/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	var (
		boshClient          *fakes.BOSHClient
		manifestGenerator   *fakes.ManifestGenerator
		logger              *fakes.Logger
		app                 compiler.Application
		compiledTempDir     string
		tempDir             string
		releaseTarballPath  string
		stemcellTarballPath string
	)

	BeforeEach(func() {
		var err error
		compiledTempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		releaseTarballPath = filepath.Join(tempDir, "some-release-42.tgz")
		err = createReleaseTarball(releaseTarballPath, bytes.NewBuffer([]byte(`---
name: some-release
version: 42
`)))
		Expect(err).NotTo(HaveOccurred())

		stemcellTarballPath = filepath.Join(tempDir, "some-stemcell-1.2.3.tgz")
		err = createStemcellTarball(stemcellTarballPath, bytes.NewBuffer([]byte(`---
operating_system: some-stemcell
version: 1.2.3
`)))
		Expect(err).NotTo(HaveOccurred())

		boshClient = &fakes.BOSHClient{}
		manifestGenerator = &fakes.ManifestGenerator{}
		logger = &fakes.Logger{}

		app = compiler.Application{
			ReleaseTarballPath:  releaseTarballPath,
			StemcellTarballPath: stemcellTarballPath,
			OutputDirectory:     compiledTempDir,
			BOSHClient:          boshClient,
			ManifestGenerator:   manifestGenerator,
			GUIDGenerator:       func() (string, error) { return "some-guid", nil },
			Logger:              logger,
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(compiledTempDir)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Run", func() {
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
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())
			Expect(boshClient.DeploymentsCall.CallCount).To(Equal(1))
			Expect(len(boshClient.DeleteDeploymentCall.Receives.Name)).To(Equal(3))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[0]).To(Equal("dep1"))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[1]).To(Equal("dep2"))
		})

		It("uploads the stemcell to the bosh director", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UploadStemcellCall.Receives.Contents).NotTo(BeNil())
			actualContents, err := ioutil.ReadAll(boshClient.UploadStemcellCall.Receives.Contents)
			Expect(err).NotTo(HaveOccurred())

			expectedContents, err := ioutil.ReadFile(stemcellTarballPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualContents).To(Equal(expectedContents))
		})

		It("uploads the release to the bosh director", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.UploadReleaseCall.Receives.Contents).NotTo(BeNil())
			actualContents, err := ioutil.ReadAll(boshClient.UploadReleaseCall.Receives.Contents)
			Expect(err).NotTo(HaveOccurred())

			expectedContents, err := ioutil.ReadFile(releaseTarballPath)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualContents).To(Equal(expectedContents))
		})

		It("generates a deployment manifest", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(manifestGenerator.GenerateCall.Receives.DirectorUUID).To(Equal("some-director-uuid"))
			Expect(manifestGenerator.GenerateCall.Receives.DeploymentName).To(Equal("compile-release-some-guid"))
			Expect(manifestGenerator.GenerateCall.Receives.Release.Name).To(Equal("some-release"))
			Expect(manifestGenerator.GenerateCall.Receives.Stemcell.Name).To(Equal("some-stemcell"))
		})

		It("deploys the manifest", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.DeployCall.Receives.Manifest).To(Equal([]byte("deployment-manifest")))
		})

		It("exports the release", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.ExportReleaseCall.Receives.DeploymentName).To(Equal("compile-release-some-guid"))
			Expect(boshClient.ExportReleaseCall.Receives.ReleaseName).To(Equal("some-release"))
			Expect(boshClient.ExportReleaseCall.Receives.ReleaseVersion).To(Equal("42"))
			Expect(boshClient.ExportReleaseCall.Receives.StemcellName).To(Equal("some-stemcell"))
			Expect(boshClient.ExportReleaseCall.Receives.StemcellVersion).To(Equal("1.2.3"))
		})

		It("downloads the compiled release", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.ResourceCall.Receives.ResourceID).To(Equal("some-resource-guid"))
		})

		It("writes the compiled release out to the given path", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			compiledReleaseContents, err := ioutil.ReadFile(filepath.Join(compiledTempDir, "some-release-42.0.0-1.2.3.tgz"))
			Expect(err).NotTo(HaveOccurred())
			Expect(compiledReleaseContents).To(Equal([]byte("compiled-release-contents")))
		})

		It("deletes the deployment", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(boshClient.DeleteDeploymentCall.Receives.Name)).To(Equal(1))
			Expect(boshClient.DeleteDeploymentCall.Receives.Name[0]).To(Equal("compile-release-some-guid"))
		})

		It("cleans up the director", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(boshClient.CleanupCall.CallCount).To(Equal(2))
		})

		It("logs all of the steps", func() {
			err := app.Run()
			Expect(err).NotTo(HaveOccurred())

			Expect(logger.Lines).To(Equal([]string{
				"deleting existing deployments\n",
				"preparing compiler\n",
				"fetching bosh director information\n",
				"generating deployment name\n",
				"parsing release details\n",
				"parsing stemcell details\n",
				"uploading stemcell some-stemcell 1.2.3\n",
				"uploading release some-release 42\n",
				"generating deployment manifest\n",
				"deploying to bosh director\n",
				"compiling the release\n",
				"downloading the compiled release\n",
				"deleting the deployment\n",
				"cleaning up\n",
			}))
		})

		Context("failure cases", func() {
			Context("when the bosh client cannot get the list of deployments", func() {
				It("returns an error", func() {
					boshClient.DeploymentsCall.Returns.Error = errors.New("failed to fetch list of deployments")

					err := app.Run()
					Expect(err).To(MatchError("failed to fetch list of deployments"))
				})
			})

			Context("when the pre-existing deployments cannot be deleted", func() {
				It("returns an error", func() {
					boshClient.DeploymentsCall.Returns.DeploymentList = []bosh.Deployment{
						{Name: "dep1"},
					}
					boshClient.DeleteDeploymentCall.Returns.Error = errors.New("failed to delete deployment")

					err := app.Run()
					Expect(err).To(MatchError("failed to delete deployment"))
				})
			})

			Context("when the bosh client cannot get the director info", func() {
				It("returns an error", func() {
					boshClient.InfoCall.Returns.Error = errors.New("failed to fetch director info")

					err := app.Run()
					Expect(err).To(MatchError("failed to fetch director info"))
				})
			})

			Context("when the guid cannot be generated", func() {
				It("returns an error", func() {
					app.GUIDGenerator = func() (string, error) { return "", errors.New("failed to generate guid") }

					err := app.Run()
					Expect(err).To(MatchError("failed to generate guid"))
				})
			})

			Context("when a release cannot be created", func() {
				It("returns an error", func() {
					app.ReleaseTarballPath = "missing-release-1.tgz"

					err := app.Run()
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when the stemcell cannot be created", func() {
				It("returns an error", func() {
					app.StemcellTarballPath = "missing-stemcell-1.tgz"

					err := app.Run()
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when the bosh client cannot upload the stemcell", func() {
				It("returns an error", func() {
					boshClient.UploadStemcellCall.Returns.Error = errors.New("failed to upload stemcell")

					err := app.Run()
					Expect(err).To(MatchError("failed to upload stemcell"))
				})
			})

			Context("when the bosh client fails to upload the release", func() {
				It("returns an error", func() {
					boshClient.UploadReleaseCall.Returns.Error = errors.New("failed to upload release")

					err := app.Run()
					Expect(err).To(MatchError("failed to upload release"))
				})
			})

			Context("when the manifest generator fails to generate the manifest", func() {
				It("returns an error", func() {
					manifestGenerator.GenerateCall.Returns.Error = errors.New("failed to generate manifest")

					err := app.Run()
					Expect(err).To(MatchError("failed to generate manifest"))
				})
			})

			Context("when the bosh client fails to deploy the manifest", func() {
				It("returns an error", func() {
					boshClient.DeployCall.Returns.Error = errors.New("failed to deploy manifest")

					err := app.Run()
					Expect(err).To(MatchError("failed to deploy manifest"))
				})
			})

			Context("when the bosh client fails to export the release", func() {
				It("returns an error", func() {
					boshClient.ExportReleaseCall.Returns.Error = errors.New("failed to export release")

					err := app.Run()
					Expect(err).To(MatchError("failed to export release"))
				})
			})

			Context("when the location for the compiled tarball cannot be written to", func() {
				AfterEach(func() {
					err := os.Chmod(compiledTempDir, 0644)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					err := os.Chmod(compiledTempDir, 0000)
					Expect(err).NotTo(HaveOccurred())

					err = app.Run()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			Context("when the bosh client cannot retrieve the resource", func() {
				It("returns an error", func() {
					boshClient.ResourceCall.Returns.Error = errors.New("failed to retrieve resource")

					err := app.Run()
					Expect(err).To(MatchError("failed to retrieve resource"))
				})
			})

			Context("when the resource cannot be written out", func() {
				var badReader *os.File

				BeforeEach(func() {
					var err error
					badReader, err = ioutil.TempFile("", "")
					Expect(err).NotTo(HaveOccurred())

					err = badReader.Close()
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					err := os.RemoveAll(badReader.Name())
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					boshClient.ResourceCall.Returns.Resource = badReader

					err := app.Run()
					Expect(err).To(MatchError(ContainSubstring("bad file descriptor")))
				})
			})

			Context("when the deployment cannot be deleted", func() {
				It("returns an error", func() {
					boshClient.DeleteDeploymentCall.Returns.Error = errors.New("failed to delete deployment")

					err := app.Run()
					Expect(err).To(MatchError("failed to delete deployment"))
				})
			})

			Context("when the bosh director cannot be cleaned up", func() {
				It("returns an error", func() {
					boshClient.CleanupCall.Returns.Error = errors.New("failed to cleanup bosh director")

					err := app.Run()
					Expect(err).To(MatchError("failed to cleanup bosh director"))
				})
			})
		})
	})
})
