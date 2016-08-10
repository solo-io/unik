package os

import (
	"archive/tar"
	log "github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"io"
	"os"
	"os/exec"
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
			continue
		}
	}

	return nil
}

///http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzip/
func Compress(source, destination string) error {
	tarCmd := exec.Command("tar", "cf", destination, "-C", source, ".")
	if out, err := tarCmd.Output(); err != nil {
		return errors.New("running tar command: "+string(out), err)
	}
	return nil
}
