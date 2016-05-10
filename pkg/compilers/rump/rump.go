package rump

import (
	"encoding/json"
	"regexp"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	uniktypes "github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/layer-x/layerx-commons/lxerrors"

	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RumpCompiler struct {
	DockerImage   string
	BakeImageName string
	CreateImage   func(kernel, args string, mntPoints []string) (*uniktypes.RawImage, error)
}

func (r *RumpCompiler) CompileRawImage(params uniktypes.CompileImageParams) (*uniktypes.RawImage, error) {
	localFolder, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return nil, err
	}

	if !params.NoCleanup {
		defer os.RemoveAll(localFolder)
	} else {
		logrus.Debugf("No cleanup for compilation. to clean, remove: " + localFolder)
	}

	logrus.Debugf("extracting uploaded files to " + localFolder)
	if err := unikos.ExtractTar(params.SourceTar, localFolder); err != nil {
		return nil, err
	}

	if err := r.runContainer(localFolder); err != nil {
		return nil, err
	}

	// now we should program compiled in local folder. next step is to bake
	progFile := path.Join(localFolder, "program")

	if !unikos.IsExists(progFile) {
		return nil, lxerrors.New("No program found - compilation failed", nil)
	}

	if err := RunContainer(r.BakeImageName, nil, []string{fmt.Sprintf("%s:%s", localFolder, "/opt/code")}, false); err != nil {
		return nil, lxerrors.New("Baking failed", err)
	}

	resultFile := path.Join(localFolder, "program.bin")

	return r.CreateImage(resultFile, params.Args, params.MntPoints)
}

func setRumpCmdLine(c rumpConfig, prog string, argv []string) rumpConfig {

	if argv == nil {
		argv = []string{}
	}

	pipe := "|"

	stub := commandLine{Bin: "stub",
		Argv: []string{},
	}
	progrc := commandLine{Bin: "program",
		Argv:    argv,
		Runmode: &pipe,
	}
	logger := commandLine{Bin: "logger",
		Argv: []string{},
	}

	c.Rc = append(c.Rc, stub, progrc, logger)
	return c
}

var netRegEx = regexp.MustCompile("net[1-9]")

// rump special json
func ToRumpJson(c rumpConfig) (string, error) {

	blk := c.Blk
	c.Blk = nil

	jsonConfig, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	blks := ""
	for _, b := range blk {

		blkjson, err := json.Marshal(b)
		if err != nil {
			return "", err
		}
		blks += fmt.Sprintf("\"blk\": %s,", string(blkjson))
	}
	var jsonString string
	if len(blks) > 0 {

		jsonString = string(jsonConfig[:len(jsonConfig)-1]) + "," + blks[:len(blks)-1] + "}"

	} else {
		jsonString = string(jsonConfig)
	}

	jsonString = netRegEx.ReplaceAllString(jsonString, "net")

	return jsonString, nil

}
