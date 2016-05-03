package osv

import (
	"io"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
	"github.com/Sirupsen/logrus"
	"io/ioutil"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"os/exec"
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
)

type OsvCompiler struct {}

func (osvCompiler *OsvCompiler) CompileRawImage(sourceTar io.ReadCloser, args string, mntPoints []string) (*types.RawImage, error) {
	localFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(localFolder)
	logrus.Debugf("extracting uploaded files to "+localFolder)
	if err := unikos.ExtractTar(sourceTar, localFolder); err != nil {
		return nil, err
	}
	cmd := exec.Command("docker", "run", "--rm", "--privileged",
		"-v", "/dev/:/dev/",
		"-v", localFolder+"/:/project_directory/",
		"projectunik/compilers-osv-java",
	)
	logrus.WithFields(logrus.Fields{
		"command": cmd.Args,
	}).Debugf("running compilers-osv-java container")
	unikutil.LogCommand(cmd, true)
	err = cmd.Run()
	if err != nil {
		return nil, lxerrors.New("failed running compilers-osv-java on "+localFolder, err)
	}
	return &types.RawImage{
		LocalImagePath: filepath.Join(localFolder, "boot.img"),
		DeviceMappings: []types.DeviceMapping{}, //TODO: not supported yet
	}, nil
}