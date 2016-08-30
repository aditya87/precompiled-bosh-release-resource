package out

import (
	"fmt"
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
	release        Release
	stemcell       Stemcell
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

func (o *OutCommand) getReleaseName() string {
	matches := regexp.MustCompile("(.*)/(.*)$").FindStringSubmatch(o.releaseDir)
	return matches[len(matches)-1]
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
		fmt.Printf("uploading stemcell %s %s\n", stemcell.Name, stemcell.Version)
		_, err = o.BOSHClient.UploadStemcell(stemcell)
		return err
	} else {
		if existingStemcell.Name == stemcell.Name && !existsInSlice(existingStemcell.Versions, stemcell.Version) {
			fmt.Printf("uploading stemcell %s %s\n", stemcell.Name, stemcell.Version)
			_, err = o.BOSHClient.UploadStemcell(stemcell)
			return err
		}
		return nil
	}
}

func (o *OutCommand) CreateRelease() error {
	createReleaseCmd := exec.Command("bosh", "create", "release", "--force", "--name", o.getReleaseName(), "--version", o.releaseVersion, "--with-tarball")
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

func (o *OutCommand) Run() error {
	fmt.Println("deleting existing deployments")
	deploymentList, err := o.BOSHClient.Deployments()
	if err != nil {
		panic(err)
	}

	for _, deployment := range deploymentList {
		err = o.BOSHClient.DeleteDeployment(o.getReleaseName())
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("fetching bosh director information")
	directorInfo, err := o.BOSHClient.Info()
	if err != nil {
		panic(err)
	}

	fmt.Println("creating release")
	err = o.CreateRelease()
	if err != nil {
		panic(err)
	}

	fmt.Println("parsing release details")
	release, err := compiler.NewRelease(filepath.Join(o.releaseDir, fmt.Sprintf("dev_releases/%s/%s-%s.tgz", o.getReleaseName(), o.getReleaseName(), o.releaseVersion)))
	if err != nil {
		panic(err)
	}

	err = o.UploadStemcell()
	if err != nil {
		panic(err)
	}

	fmt.Printf("uploading release %s %s\n", release.Name, release.Version)
	_, err = o.BOSHClient.UploadRelease(release)
	if err != nil {
		panic(err)
	}

	fmt.Println("generating deployment manifest")
	manifest, err := compiler.ManifestGenerator.Generate(directorInfo.UUID, release.Name, release, stemcell)
	if err != nil {
		panic(err)
	}

	fmt.Println("deploying to bosh director")
	_, err = o.BOSHClient.Deploy(manifest)
	if err != nil {
		panic(err)
	}

	fmt.Println("compiling the release")
	resourceID, err := o.BOSHClient.ExportRelease(deploymentName, release.Name, release.Version, stemcell.Name, stemcell.Version)
	if err != nil {
		panic(err)
	}

	fmt.Println("downloading the compiled release")
	compiledTarballPath := filepath.Join(a.OutputDirectory, fmt.Sprintf("%s-%s-%s.tgz", release.Name, release.Semver, stemcell.Semver))
	fd, err := os.OpenFile(compiledTarballPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	resource, err := o.BOSHClient.Resource(resourceID)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(fd, resource)
	if err != nil {
		panic(err)
	}

	fmt.Println("deleting the deployment")
	err = o.BOSHClient.DeleteDeployment(deploymentName)
	if err != nil {
		panic(err)
	}

	fmt.Println("cleaning up")
	_, err = o.BOSHClient.Cleanup()
	if err != nil {
		panic(err)
	}

	return nil
}
