package compiler_test

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/pcf-releng-ci/tasks/future/compile-release/compiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Release", func() {
	Describe("NewRelease", func() {
		var tempDir string

		BeforeEach(func() {
			var err error
			tempDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when parsing version numbers", func() {
			It("can parse a one-part version number", func() {
				path := filepath.Join(tempDir, "release.tgz")
				err := createReleaseTarball(path, bytes.NewBuffer([]byte(`---
name: one-part-version
version: 1
`)))

				release, err := compiler.NewRelease(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Name).To(Equal("one-part-version"))
				Expect(release.Semver).To(Equal(compiler.Semver{
					Major: 1,
				}))
			})

			It("can parse a two-part version number", func() {
				path := filepath.Join(tempDir, "release.tgz")
				err := createReleaseTarball(path, bytes.NewBuffer([]byte(`---
name: two-part-version
version: 1.2
`)))

				release, err := compiler.NewRelease(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Name).To(Equal("two-part-version"))
				Expect(release.Semver).To(Equal(compiler.Semver{
					Major: 1,
					Minor: 2,
				}))
			})

			It("can parse a three-part version number", func() {
				path := filepath.Join(tempDir, "release.tgz")
				err := createReleaseTarball(path, bytes.NewBuffer([]byte(`---
name: three-part-version
version: 1.2.3
`)))

				release, err := compiler.NewRelease(path)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Name).To(Equal("three-part-version"))
				Expect(release.Semver).To(Equal(compiler.Semver{
					Major: 1,
					Minor: 2,
					Patch: 3,
				}))
			})
		})

		Context("failure cases", func() {
			Context("when the release tarball does not exist", func() {
				It("returns an error", func() {
					_, err := compiler.NewRelease("some-release-1.tgz")
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})

			Context("when the release tarball is not gzipped", func() {
				It("returns an error", func() {
					path := filepath.Join(tempDir, "release.tgz")
					err := ioutil.WriteFile(path, []byte("not gzipped"), 0644)
					Expect(err).NotTo(HaveOccurred())

					_, err = compiler.NewRelease(path)
					Expect(err).To(MatchError("gzip: invalid header"))
				})
			})

			Context("when the header file is corrupt", func() {
				It("returns an error", func() {
					path := filepath.Join(tempDir, "release.tgz")
					tarball, err := os.Create(path)
					Expect(err).NotTo(HaveOccurred())

					gw := gzip.NewWriter(tarball)
					tw := tar.NewWriter(gw)

					err = tw.Close()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					_, err = compiler.NewRelease(path)
					Expect(err).To(MatchError(fmt.Sprintf("could not find release.MF in %q", path)))
				})
			})

			Context("when there is no release.MF", func() {
				It("returns an error", func() {
					path := filepath.Join(tempDir, "release.tgz")
					tarball, err := os.Create(path)
					Expect(err).NotTo(HaveOccurred())

					gw := gzip.NewWriter(tarball)
					tw := tar.NewWriter(gw)

					releaseManifest := bytes.NewBuffer([]byte(`---
name: release
version: 1.2.3
`))

					header := &tar.Header{
						Name:    "./someotherfile.MF",
						Size:    int64(releaseManifest.Len()),
						Mode:    int64(0644),
						ModTime: time.Now(),
					}

					err = tw.WriteHeader(header)
					Expect(err).NotTo(HaveOccurred())

					_, err = io.Copy(tw, releaseManifest)
					Expect(err).NotTo(HaveOccurred())

					err = tw.Close()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					_, err = compiler.NewRelease(path)
					Expect(err).To(MatchError(fmt.Sprintf("could not find release.MF in %q", path)))
				})
			})

			Context("when the tarball is corrupt", func() {
				It("returns an error", func() {
					path := filepath.Join(tempDir, "release.tgz")
					tarball, err := os.Create(path)
					Expect(err).NotTo(HaveOccurred())

					gw := gzip.NewWriter(tarball)
					tw := bufio.NewWriter(gw)

					_, err = tw.WriteString("I am a banana!")
					Expect(err).NotTo(HaveOccurred())

					err = tw.Flush()
					Expect(err).NotTo(HaveOccurred())

					err = gw.Close()
					Expect(err).NotTo(HaveOccurred())

					_, err = compiler.NewRelease(path)
					Expect(err).To(MatchError(fmt.Sprintf("error while reading %q: unexpected EOF", path)))
				})
			})

			Context("when the release manifest is not YAML", func() {
				It("returns an error", func() {
					path := filepath.Join(tempDir, "release.tgz")
					err := createReleaseTarball(path, bytes.NewBuffer([]byte("%%%%%")))
					Expect(err).NotTo(HaveOccurred())

					_, err = compiler.NewRelease(path)
					Expect(err).To(MatchError("yaml: could not find expected directive name"))
				})
			})
		})
	})
})
