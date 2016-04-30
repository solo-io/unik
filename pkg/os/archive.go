package os

import (
	"archive/tar"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"os"
	"path"
)

func ExtractTar(tarArchive io.ReadCloser, localFolder string) error {
	tr := tar.NewReader(tarArchive)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}
		log.WithField("file", hdr.Name).Debug("Extracting file")
		switch hdr.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(path.Join(localFolder, hdr.Name), 0755)

			if err != nil {
				return err
			}

		case tar.TypeReg:
			fallthrough
		case tar.TypeRegA:
			dir, _ := path.Split(hdr.Name)
			if err := os.MkdirAll(path.Join(localFolder, dir), 0755); err != nil {
				return err
			}

			outputFile, err := os.Create(path.Join(localFolder, hdr.Name))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outputFile, tr); err != nil {
				outputFile.Close()
				return err
			}
			outputFile.Close()

		default:
			return errors.New("Unsupported file type in tar")
		}
	}

	return nil
}

///http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzip/
func Compress(source, destination string) error {
	tarfile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer tarfile.Close()

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.Join(filepath.Base(baseDir), strings.TrimPrefix(path, source))
			}

			header.Name = filepath.Base(path)

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}