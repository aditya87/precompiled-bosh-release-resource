package out

import (
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/aditya87/precompiled-bosh-release-resource/compiler"
	"github.com/pivotal-cf-experimental/bosh-test/bosh"
)

type OutCommand struct {
	BOSHClient     boshClient
	releaseDir     string
	releaseVersion string
	stemcellDir    string
}

type boshClient interface {
	Resource(resourceID string) (file io.ReadCloser, err error)
	ExportRelease(deploymentName, releaseName, releaseVersion, stemcellName, stemcellVersion string) (resourceID string, err error)
	Deploy(manifest []byte) (taskID int, err error)
	UploadStemcell(stemcell bosh.SizeReader) (taskID int, err error)
	UploadRelease(release bosh.SizeReader) (taskID int, err error)
	Info() (bosh.DirectorInfo, error)
	DeleteDeployment(name string) error
	Cleanup() (taskID int, err error)
	Deployments() (deploymentList []bosh.Deployment, err error)
}

type manifestGenerator interface {
	Generate(directorUUID, deploymentName string, release compiler.Release, stemcell compiler.Stemcell) (manifest []byte, err error)
}

func NewOutCommand(request OutRequest) *OutCommand {
	return &OutCommand{
		BOSHClient: bosh.NewClient(bosh.Config{
			URL:              request.Source.BoshTarget,
			Username:         request.Source.BoshUser,
			Password:         request.Source.BoshPassword,
			AllowInsecureSSL: true,
		}),
		releaseDir:     request.Params.ReleaseDir,
		releaseVersion: request.Params.ReleaseVersion,
		stemcellDir:    request.Params.StemcellDir,
	}
}

func (o *OutCommand) UploadStemcell() error {
	stemcellDirInfo, err := ioutil.ReadDir(o.stemcellDir)
	if err != nil {
		panic(err)
	}
	stemcellTarballPath := filepath.Join(o.stemcellDir, stemcellDirInfo[0].Name())
	stemcell, err := compiler.NewStemcell(stemcellTarballPath)
	if err != nil {
		panic(err)
	}

	_, err = o.BOSHClient.UploadStemcell(stemcell)
	return err
}
