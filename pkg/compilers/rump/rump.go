package rump

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

// uses rump docker conter container
// the container expectes code in /opt/code and will produce program.bin in the same folder.
// we need to take the program bin and combine with json config produce an image

type RumpGoCompiler struct {
	RumCompilerBase
}

func (r *RumpGoCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {
	sourcesDir := params.SourcesDir
	godepsFile := filepath.Join(sourcesDir, "Godeps", "Godeps.json")
	_, err := os.Stat(godepsFile)
	if err != nil {
		return nil, errors.New("the Go compiler requires Godeps file in the root of your project. see https://github.com/tools/godep", nil)
	}
	data, err := ioutil.ReadFile(godepsFile)
	if err != nil {
		return nil, errors.New("could not read godeps file", err)
	}
	var g godeps
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, errors.New("invalid json in godeps file", err)
	}
	containerEnv := []string{
		fmt.Sprintf("ROOT_PATH=%s", g.ImportPath),
	}

	// should we use the baker stubs?
	if r.BakeImageName != "" {

		if err := r.runAndBake(sourcesDir, containerEnv); err != nil {
			return nil, err
		}
	} else {

		if err := r.runContainer(sourcesDir, containerEnv); err != nil {
			return nil, err
		}
	}
	// now we should program.bin
	resultFile := path.Join(sourcesDir, "program.bin")
	logrus.Debugf("finished kernel binary at %s", resultFile)
	img, err := r.CreateImage(resultFile, params.Args, params.MntPoints, nil, params.NoCleanup)
	if err != nil {
		return nil, errors.New("creating boot volume from kernel binary", err)
	}
	return img, nil
}

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

	return jsonString, nil

}

type godeps struct {
	ImportPath   string   `json:"ImportPath"`
	GoVersion    string   `json:"GoVersion"`
	GodepVersion string   `json:"GodepVersion"`
	Packages     []string `json:"Packages"`
	Deps         []struct {
		ImportPath string `json:"ImportPath"`
		Rev        string `json:"Rev"`
		Comment    string `json:"Comment,omitempty"`
	} `json:"Deps"`
}
