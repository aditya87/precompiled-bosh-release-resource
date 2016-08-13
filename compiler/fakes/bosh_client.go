package fakes

import (
	"io"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"
)

type BOSHClient struct {
	InfoCall struct {
		CallCount int
		Returns   struct {
			DirectorInfo bosh.DirectorInfo
			Error        error
		}
	}

	UploadReleaseCall struct {
		Receives struct {
			Contents bosh.SizeReader
		}
		Returns struct {
			TaskID int
			Error  error
		}
	}

	UploadStemcellCall struct {
		CallCount int
		Receives  struct {
			Contents bosh.SizeReader
		}
		Returns struct {
			TaskID int
			Error  error
		}
	}

	DeployCall struct {
		Receives struct {
			Manifest []byte
		}
		Returns struct {
			TaskID int
			Error  error
		}
	}

	ExportReleaseCall struct {
		Receives struct {
			DeploymentName  string
			ReleaseName     string
			ReleaseVersion  string
			StemcellName    string
			StemcellVersion string
		}
		Returns struct {
			ResourceID string
			Error      error
		}
	}

	ResourceCall struct {
		Receives struct {
			ResourceID string
		}
		Returns struct {
			Resource io.ReadCloser
			Error    error
		}
	}

	DeleteDeploymentCall struct {
		Receives struct {
			Name []string
		}
		Returns struct {
			Error error
		}
	}

	CleanupCall struct {
		CallCount int
		Returns   struct {
			TaskID int
			Error  error
		}
	}

	DeploymentsCall struct {
		CallCount int
		Returns   struct {
			DeploymentList []bosh.Deployment
			Error          error
		}
	}

	StemcellCall struct {
		CallCount int
		Receives  string
		Returns   struct {
			Stemcell bosh.Stemcell
			Error    error
		}
	}
}

func (c *BOSHClient) Info() (bosh.DirectorInfo, error) {
	c.InfoCall.CallCount++

	return c.InfoCall.Returns.DirectorInfo, c.InfoCall.Returns.Error
}

func (c *BOSHClient) Deploy(manifest []byte) (int, error) {
	c.DeployCall.Receives.Manifest = manifest

	return c.DeployCall.Returns.TaskID, c.DeployCall.Returns.Error
}

func (c *BOSHClient) ExportRelease(deploymentName, releaseName, releaseVersion, stemcellName, stemcellVersion string) (string, error) {
	c.ExportReleaseCall.Receives.DeploymentName = deploymentName
	c.ExportReleaseCall.Receives.ReleaseName = releaseName
	c.ExportReleaseCall.Receives.ReleaseVersion = releaseVersion
	c.ExportReleaseCall.Receives.StemcellName = stemcellName
	c.ExportReleaseCall.Receives.StemcellVersion = stemcellVersion

	return c.ExportReleaseCall.Returns.ResourceID, c.ExportReleaseCall.Returns.Error
}

func (c *BOSHClient) Resource(resourceID string) (io.ReadCloser, error) {
	c.ResourceCall.Receives.ResourceID = resourceID

	return c.ResourceCall.Returns.Resource, c.ResourceCall.Returns.Error
}

func (c *BOSHClient) UploadRelease(contents bosh.SizeReader) (int, error) {
	c.UploadReleaseCall.Receives.Contents = contents

	return c.UploadReleaseCall.Returns.TaskID, c.UploadReleaseCall.Returns.Error
}

func (c *BOSHClient) UploadStemcell(contents bosh.SizeReader) (int, error) {
	c.UploadStemcellCall.CallCount++
	c.UploadStemcellCall.Receives.Contents = contents

	return c.UploadStemcellCall.Returns.TaskID, c.UploadStemcellCall.Returns.Error
}

func (c *BOSHClient) DeleteDeployment(name string) error {
	c.DeleteDeploymentCall.Receives.Name = append(c.DeleteDeploymentCall.Receives.Name, name)

	return c.DeleteDeploymentCall.Returns.Error
}

func (c *BOSHClient) Cleanup() (int, error) {
	c.CleanupCall.CallCount++

	return c.CleanupCall.Returns.TaskID, c.CleanupCall.Returns.Error
}

func (c *BOSHClient) Deployments() ([]bosh.Deployment, error) {
	c.DeploymentsCall.CallCount++

	return c.DeploymentsCall.Returns.DeploymentList, c.DeploymentsCall.Returns.Error
}

func (c *BOSHClient) Stemcell(name string) (bosh.Stemcell, error) {
	c.StemcellCall.CallCount++
	c.StemcellCall.Receives = name

	return c.StemcellCall.Returns.Stemcell, c.StemcellCall.Returns.Error
}
