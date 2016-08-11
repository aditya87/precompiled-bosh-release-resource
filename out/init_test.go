package out_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestOut(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Out Suite")
}

func createReleaseTarball(path string, manifest *bytes.Buffer) error {
	tarball, err := os.Create(path)
	if err != nil {
		return err
	}
	defer tarball.Close()

	gw := gzip.NewWriter(tarball)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	header := &tar.Header{
		Name:    "./release.MF",
		Size:    int64(manifest.Len()),
		Mode:    int64(0644),
		ModTime: time.Now(),
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, manifest)
	if err != nil {
		return err
	}

	return nil
}

func createStemcellTarball(path string, manifest *bytes.Buffer) error {
	tarball, err := os.Create(path)
	if err != nil {
		return err
	}
	defer tarball.Close()

	gw := gzip.NewWriter(tarball)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	header := &tar.Header{
		Name:    "./stemcell.MF",
		Size:    int64(manifest.Len()),
		Mode:    int64(0644),
		ModTime: time.Now(),
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, manifest)
	if err != nil {
		return err
	}

	return nil
}
