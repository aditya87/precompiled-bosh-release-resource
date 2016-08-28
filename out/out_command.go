package out

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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
	Stemcell(name string) (bosh.Stemcell, error)
}

type manifestGenerator interface {
	Generate(directorUUID, deploymentName string, release compiler.Release, stemcell compiler.Stemcell) (manifest []byte, err error)
}

func existsInSlice(slice []string, str string) bool {
	for _, x := range slice {
		if x == str {
			return true
		}
	}
	return false
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

	existingStemcell, err := o.BOSHClient.Stemcell(stemcell.Name)
	if err != nil && !strings.Contains(err.Error(), "could not be found") {
		panic(err)
	} else if err != nil && strings.Contains(err.Error(), "could not be found") {
		_, err = o.BOSHClient.UploadStemcell(stemcell)
		return err
	} else {
		if existingStemcell.Name == stemcell.Name && !existsInSlice(existingStemcell.Versions, stemcell.Version) {
			_, err = o.BOSHClient.UploadStemcell(stemcell)
			return err
		}
		return nil
	}
}

func (o *OutCommand) CreateRelease() error {
	matches := regexp.MustCompile("(.*)/(.*)$").FindStringSubmatch(o.releaseDir)
	createReleaseCmd := exec.Command("bosh", "create", "release", "--force", "--name", matches[len(matches)-1], "--version", o.releaseVersion, "--with-tarball")
	createReleaseCmd.Dir = o.releaseDir
	err := os.RemoveAll(filepath.Join(o.releaseDir, "dev_releases"))
	if err != nil {
		panic(err)
	}

	err = createReleaseCmd.Run()
	if err != nil {
		panic(err)
	}
	return nil
}
