package compilers

import (
	"archive/tar"
	"io"

	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"

	log "github.com/Sirupsen/logrus"

	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RunmpCompiler struct {
	DockerImage string
	CreateImage func(kernel, args string, mntPoints []string) (*uniktypes.RawImage, error)
}

func (r *RunmpCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*uniktypes.RawImage, error) {

	localFolder, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)

	if err := r.extractTar(sourceTar, localFolder); err != nil {
		return nil, err
	}

	if err := r.runContainer(localFolder); err != nil {
		return nil, err
	}

	// now we should program.bin
	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, args, mntPoints)
}

func (r *RunmpCompiler) extractTar(tarArchive io.ReadCloser, localFolder string) error {
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

func (r *RunmpCompiler) runContainer(localFolder string) error {

	return RunContainer(r.DockerImage, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false)
}
