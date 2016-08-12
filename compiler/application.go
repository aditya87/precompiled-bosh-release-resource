package compiler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
)

type Application struct {
	ReleaseTarballPath  string
	StemcellTarballPath string
	OutputDirectory     string
	BOSHClient          boshClient
	ManifestGenerator   manifestGenerator
	GUIDGenerator       func() (string, error)
	Logger              logger
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
	Generate(directorUUID, deploymentName string, release Release, stemcell Stemcell) (manifest []byte, err error)
}

type logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

func (a Application) Run() error {
	a.Logger.Println("deleting existing deployments")
	deploymentList, err := a.BOSHClient.Deployments()
	if err != nil {
		return err
	}

	for _, deployment := range deploymentList {
		err = a.BOSHClient.DeleteDeployment(deployment.Name)
		if err != nil {
			return err
		}
	}

	a.Logger.Println("preparing compiler")
	_, err = a.BOSHClient.Cleanup()
	if err != nil {
		return err
	}

	a.Logger.Println("fetching bosh director information")
	directorInfo, err := a.BOSHClient.Info()
	if err != nil {
		return err
	}

	a.Logger.Println("generating deployment name")
	guid, err := a.GUIDGenerator()
	if err != nil {
		return err
	}

	deploymentName := fmt.Sprintf("compile-release-%s", guid)

	a.Logger.Println("parsing release details")
	release, err := NewRelease(a.ReleaseTarballPath)
	if err != nil {
		return err
	}

	a.Logger.Println("parsing stemcell details")
	stemcell, err := NewStemcell(a.StemcellTarballPath)
	if err != nil {
		return err
	}

	a.Logger.Printf("uploading stemcell %s %s\n", stemcell.Name, stemcell.Version)
	_, err = a.BOSHClient.UploadStemcell(stemcell)
	if err != nil {
		return err
	}

	a.Logger.Printf("uploading release %s %s\n", release.Name, release.Version)
	_, err = a.BOSHClient.UploadRelease(release)
	if err != nil {
		return err
	}

	a.Logger.Println("generating deployment manifest")
	manifest, err := a.ManifestGenerator.Generate(directorInfo.UUID, deploymentName, release, stemcell)
	if err != nil {
		return err
	}

	a.Logger.Println("deploying to bosh director")
	_, err = a.BOSHClient.Deploy(manifest)
	if err != nil {
		return err
	}

	a.Logger.Println("compiling the release")
	resourceID, err := a.BOSHClient.ExportRelease(deploymentName, release.Name, release.Version, stemcell.Name, stemcell.Version)
	if err != nil {
		return err
	}

	a.Logger.Println("downloading the compiled release")
	compiledTarballPath := filepath.Join(a.OutputDirectory, fmt.Sprintf("%s-%s-%s.tgz", release.Name, release.Semver, stemcell.Semver))
	fd, err := os.OpenFile(compiledTarballPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	resource, err := a.BOSHClient.Resource(resourceID)
	if err != nil {
		return err
	}

	_, err = io.Copy(fd, resource)
	if err != nil {
		return err
	}

	a.Logger.Println("deleting the deployment")
	err = a.BOSHClient.DeleteDeployment(deploymentName)
	if err != nil {
		return err
	}

	a.Logger.Println("cleaning up")
	_, err = a.BOSHClient.Cleanup()
	if err != nil {
		return err
	}

	return nil
}
