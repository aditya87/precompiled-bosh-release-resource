package compiler_test

import (
	"github.com/pivotal-cf/pcf-releng-ci/tasks/future/compile-release/compiler"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("ManifestGenerator", func() {
	It("creates a manifest with the given releases and stemcell", func() {
		generator := compiler.NewManifestGenerator()
		manifest, err := generator.Generate("some-director-uuid", "compiled-release-guid", compiler.Release{
			Name:    "some-release-name-1",
			Version: "some-release-version-1",
		}, compiler.Stemcell{
			Name:    "some-stemcell-os",
			Version: "some-stemcell-version",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(manifest).To(MatchYAML(`---
name: compiled-release-guid
director_uuid: some-director-uuid
releases:
  - name: some-release-name-1
    version: some-release-version-1
stemcells:
  - alias: default
    os: some-stemcell-os
    version: some-stemcell-version
update:
  canaries: 1
  max_in_flight: 1
  canary_watch_time: 1000-1001
  update_watch_time: 1000-1001
instance_groups: []
`))
	})
})
