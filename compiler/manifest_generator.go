package compiler

import "gopkg.in/yaml.v2"

type guidGenerator func() string

type ManifestGenerator struct {
	guidGenerator guidGenerator
}

type Manifest struct {
	Name           string             `yaml:"name"`
	DirectorUUID   string             `yaml:"director_uuid"`
	Releases       []ManifestRelease  `yaml:"releases"`
	Stemcells      []ManifestStemcell `yaml:"stemcells"`
	Update         ManifestUpdate     `yaml:"update"`
	InstanceGroups []interface{}      `yaml:"instance_groups"`
}

type ManifestRelease struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ManifestStemcell struct {
	Alias   string `yaml:"alias"`
	OS      string `yaml:"os"`
	Version string `yaml:"version"`
}

type ManifestUpdate struct {
	Canaries        int    `yaml:"canaries"`
	MaxInFlight     int    `yaml:"max_in_flight"`
	CanaryWatchTime string `yaml:"canary_watch_time"`
	UpdateWatchTime string `yaml:"update_watch_time"`
}

func NewManifestGenerator() ManifestGenerator {
	return ManifestGenerator{}
}

func (g ManifestGenerator) Generate(directorUUID, deploymentName string, release Release, stemcell Stemcell) ([]byte, error) {
	manifest := Manifest{
		Name:         deploymentName,
		DirectorUUID: directorUUID,
		Releases:     []ManifestRelease{},
		Stemcells:    []ManifestStemcell{},
		Update: ManifestUpdate{
			Canaries:        1,
			MaxInFlight:     1,
			CanaryWatchTime: "1000-1001",
			UpdateWatchTime: "1000-1001",
		},
	}

	manifest.Releases = append(
		manifest.Releases,
		ManifestRelease{
			Name:    release.Name,
			Version: release.Version,
		},
	)

	manifest.Stemcells = append(
		manifest.Stemcells,
		ManifestStemcell{
			Alias:   "default",
			OS:      stemcell.Name,
			Version: stemcell.Version,
		},
	)

	manifestYAML, err := yaml.Marshal(manifest)
	if err != nil {
		return nil, err
	}

	return manifestYAML, nil
}
