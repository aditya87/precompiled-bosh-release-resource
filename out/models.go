package out

import "github.com/aditya87/precompiled-bosh-release-resource"

type OutRequest struct {
	Source precompiled_release_resource.Source `json:"source"`
	Params Params                              `json:"params"`
}

type Params struct {
	ReleaseDir     string `json:"release_dir"`
	ReleaseVersion string `json:"release_version"`
	StemcellDir    string `json:"stemcell_dir"`
}
