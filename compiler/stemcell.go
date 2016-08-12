package compiler

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

var tarballRegex = regexp.MustCompile(`(.*)-([\d\.]+)\.tgz`)

type Stemcell struct {
	Name    string `yaml:"operating_system"`
	Version string
	Semver  Semver
	*os.File
	size int64
}

func NewStemcell(path string) (Stemcell, error) {
	fd, err := os.Open(path)
	if err != nil {
		return Stemcell{}, err
	}
	defer fd.Close()

	fileInfo, err := fd.Stat()
	if err != nil {
		return Stemcell{}, err
	}

	gr, err := gzip.NewReader(fd)
	if err != nil {
		return Stemcell{}, err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	header, err := tr.Next()
	for err == nil {
		if filepath.Base(header.Name) == "stemcell.MF" {
			break
		}

		header, err = tr.Next()
	}
	if err != nil {
		if err == io.EOF {
			return Stemcell{}, fmt.Errorf("could not find stemcell.MF in %q", path)
		}

		return Stemcell{}, fmt.Errorf("error while reading %q: %s", path, err)
	}

	var stemcell Stemcell
	content, err := ioutil.ReadAll(tr)
	if err != nil {
		return Stemcell{}, err
	}

	err = yaml.Unmarshal(content, &stemcell)
	if err != nil {
		return Stemcell{}, err
	}

	stemcell.File, err = os.Open(path)
	if err != nil {
		return Stemcell{}, err
	}

	stemcell.size = fileInfo.Size()

	parts := strings.Split(stemcell.Version, ".")
	switch len(parts) {
	case 1:
		stemcell.Semver.Major, _ = strconv.Atoi(parts[0])
	case 2:
		stemcell.Semver.Major, _ = strconv.Atoi(parts[0])
		stemcell.Semver.Minor, _ = strconv.Atoi(parts[1])
	case 3:
		stemcell.Semver.Major, _ = strconv.Atoi(parts[0])
		stemcell.Semver.Minor, _ = strconv.Atoi(parts[1])
		stemcell.Semver.Patch, _ = strconv.Atoi(parts[2])
	default:
		panic(err)
	}

	return stemcell, nil
}

func (s Stemcell) Size() int64 {
	return s.size
}
