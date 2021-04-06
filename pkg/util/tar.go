package util

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func DecompressToFile(r io.Reader, destPath string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		filename := filepath.Join(destPath, hdr.Name)
		file, err := createFile(filename)
		if err != nil {
			return err
		}
		io.Copy(file, tr)
	}
	return nil
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(name), 0755)
	if err != nil {
		return nil, err
	}
	return os.Create(name)
}

func tarFromReader(r io.Reader, name string, size int64, tw *tar.Writer) error {
	if err := tw.WriteHeader(&tar.Header{
		Mode: 0644,
		Size: size,
		Name: name,
	}); err != nil {
		return err
	}

	_, err := io.Copy(tw, r)
	if err != nil {
		return err
	}

	return nil
}
