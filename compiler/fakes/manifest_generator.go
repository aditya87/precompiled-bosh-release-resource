package fakes

import "github.com/pivotal-cf/pcf-releng-ci/tasks/future/compile-release/compiler"

type ManifestGenerator struct {
	GenerateCall struct {
		Receives struct {
			DeploymentName string
			DirectorUUID   string
			Stemcell       compiler.Stemcell
			Release        compiler.Release
		}
		Returns struct {
			Manifest []byte
			Error    error
		}
	}
}

func (g *ManifestGenerator) Generate(directorUUID, deploymentName string, release compiler.Release, stemcell compiler.Stemcell) ([]byte, error) {
	g.GenerateCall.Receives.DirectorUUID = directorUUID
	g.GenerateCall.Receives.Release = release
	g.GenerateCall.Receives.Stemcell = stemcell
	g.GenerateCall.Receives.DeploymentName = deploymentName

	return g.GenerateCall.Returns.Manifest, g.GenerateCall.Returns.Error
}
