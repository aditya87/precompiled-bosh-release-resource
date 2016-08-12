package compiler

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Semver struct {
	Major int
	Minor int
	Patch int
}

func (s Semver) String() string {
	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

type Release struct {
	Name    string
	Version string
	Semver  Semver
	*os.File
	size int64
}

func NewRelease(path string) (Release, error) {
	fd, err := os.Open(path)
	if err != nil {
		return Release{}, err
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		return Release{}, err
	}

	gr, err := gzip.NewReader(fd)
	if err != nil {
		return Release{}, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	header, err := tr.Next()
	for err == nil {
		if filepath.Base(header.Name) == "release.MF" {
			break
		}

		header, err = tr.Next()
	}
	if err != nil {
		if err == io.EOF {
			return Release{}, fmt.Errorf("could not find release.MF in %q", path)
		}

		return Release{}, fmt.Errorf("error while reading %q: %s", path, err)
	}

	var release Release
	content, err := ioutil.ReadAll(tr)
	if err != nil {
		return Release{}, err
	}

	err = yaml.Unmarshal(content, &release)
	if err != nil {
		return Release{}, err
	}

	release.File, err = os.Open(path)
	if err != nil {
		return Release{}, err
	}

	release.size = fileInfo.Size()

	parts := strings.Split(release.Version, ".")
	switch len(parts) {
	case 1:
		release.Semver.Major, _ = strconv.Atoi(parts[0])
	case 2:
		release.Semver.Major, _ = strconv.Atoi(parts[0])
		release.Semver.Minor, _ = strconv.Atoi(parts[1])
	case 3:
		release.Semver.Major, _ = strconv.Atoi(parts[0])
		release.Semver.Minor, _ = strconv.Atoi(parts[1])
		release.Semver.Patch, _ = strconv.Atoi(parts[2])
	default:
		return Release{}, fmt.Errorf("could not parse semver version from %s", release.Version)
	}

	return release, nil
}

func (r Release) Size() int64 {
	return r.size
}
