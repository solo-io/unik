package mirage

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"

	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"gopkg.in/yaml.v2"
)

/*
plan:
# pass args from manifest from parsed by unik
# unik will then read back the xl file and parse it
# will search for *img files and use them as volumes for the unikernel (mountpoints per say)
# done!

*/
type MirageCompiler struct {
}

func (c *MirageCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {

	sourcesDir := params.SourcesDir

	if err := grantOpamPermissions(sourcesDir); err != nil {
		return nil, err
	}

	args, err := parseMirageManifest(sourcesDir)
	if err != nil {
		args, err = introspectArguments(sourcesDir)
		if err != nil {
			return nil, err
		}
	}

	args = append([]string{"configure", "-t", "xen"}, args...)
	if err := unikutil.NewContainer("compilers-mirage-ocaml-xen").WithEntrypoint("mirage").WithVolume(sourcesDir, "/opt/code").Run(args...); err != nil {
		return nil, err
	}

	if err := unikutil.NewContainer("compilers-mirage-ocaml-xen").WithEntrypoint("/usr/bin/make").WithVolume(sourcesDir, "/opt/code").Run(); err != nil {
		return nil, err
	}

	// Extract volume info

	// read xl template file, and see what disks are needed

	matches, err := filepath.Glob(filepath.Join(params.SourcesDir, "*.xl.in"))
	if err != nil {
		return nil, err
	}

	if len(matches) != 1 {
		return nil, errors.New("XL file count is wrong", nil)
	}

	xlFile := matches[0]

	disks, err := getDisks(sourcesDir, xlFile)
	if err != nil {
		return nil, err
	}

	var res types.RawImage
	unikernelfile, err := getUnikernelFile(sourcesDir)
	if err != nil {
		return nil, err
	}

	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"})
	for _, disk := range disks {
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "xen:" + disk, DeviceName: disk})
	}

	imgFile, err := compilers.BuildBootableImage(unikernelfile, "", false, params.NoCleanup)

	if err != nil {
		return nil, err
	}

	res.LocalImagePath = imgFile
	res.StageSpec = types.StageSpec{
		ImageFormat:           types.ImageFormat_RAW,
		XenVirtualizationType: types.XenVirtualizationType_Paravirtual,
	}
	res.RunSpec.DefaultInstanceMemory = 256

	return &res, nil
}

var parseRegEx = regexp.MustCompile(`vdev=(\S+),\s.+?target=@\S+?:(\S+?)@`)

func getUnikernelFile(sourcesDir string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(sourcesDir, "*.xen"))
	if err != nil {
		return "", nil
	}

	if len(matches) != 1 {
		return "", errors.New("Xen kernel file count is wrong", nil)
	}

	return matches[0], nil
}

func getDisks(sourcesDir string, xlFile string) ([]string, error) {

	xlFileFile, err := os.Open(xlFile)
	if err != nil {
		return nil, err
	}
	defer xlFileFile.Close()
	return getDisksFromReader(sourcesDir, xlFileFile)
}

func getDiskMatchesFromReader(xlFile io.Reader) ([][]string, error) {

	scanner := bufio.NewScanner(xlFile)
	const diskPrefix = "disk = "
	var matches [][]string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, diskPrefix) {
			restOfLine := line[len(diskPrefix):]
			matches = parseRegEx.FindAllStringSubmatch(restOfLine, -1)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil

}

func getDisksFromReader(sourcesDir string, xlFile io.Reader) ([]string, error) {
	matches, err := getDiskMatchesFromReader(xlFile)

	if err != nil {
		return nil, err
	}

	var disks []string
	for _, match := range matches {
		// match must have length of two!
		if len(match) != 3 {
			return nil, errors.New("Matches mismatch - please update code", nil)
		}
		disk := match[1]
		volFile := match[2]

		_, err := os.Stat(filepath.Join(sourcesDir, volFile))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				return nil, errors.New("unexpected error statting disk file", err)
			}
		}
		disks = append(disks, disk)
	}

	// only return disks that have an image file with them...

	return disks, nil
}

func introspectArguments(sourcesDir string) ([]string, error) {

	output, err := unikutil.NewContainer("compilers-mirage-ocaml-xen").WithVolume(sourcesDir, "/opt/code").WithEntrypoint("/home/opam/.opam/system/bin/mirage").CombinedOutput("describe", "--color=never")
	if err != nil {
		log.WithError(err).WithFields(log.Fields{"output": string(output)}).Error("Error getting data on mirage code")
		return nil, err
	}

	keys := getKeys(getKeyStringFromDescribe(string(output)))

	var args []string

	if _, ok := keys["kv_ro"]; ok {
		args = append(args, "--kv_ro", "fat")
	}

	if _, ok := keys["network"]; ok {
		args = append(args, "--network", "0")
		args = append(args, "--net", "direct")
		args = append(args, "--dhcp", "true")
	}

	// find kv_ro, net, network arguments
	return args, nil
}

func getKeyStringFromDescribe(input string) string {
	const keys = "Keys "
	index := strings.Index(input, keys)

	if index < 0 {
		return ""
	}
	return input[index+len(keys):]
}

var pairs = regexp.MustCompile(`\s*(\S+?)=(.+?)(,|\s*$)`)

func getKeys(input string) map[string]string {
	res := make(map[string]string)
	matches := pairs.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		res[match[1]] = match[2]
	}
	return res
}

type mirageProjectConfig struct {
	Args string `yaml:"arguments"`
}

func parseMirageManifest(sourcesDir string) ([]string, error) {
	data, err := ioutil.ReadFile(filepath.Join(sourcesDir, "manifest.yaml"))
	if err != nil {
		return nil, errors.New("failed to read manifest.yaml file", err)
	}

	var config mirageProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.New("failed to parse yaml manifest.yaml file", err)
	}

	return strings.Split(config.Args, " "), nil
}

func grantOpamPermissions(sourcesDir string) error {
	err := unikutil.NewContainer("compilers-mirage-ocaml-xen").WithVolume(sourcesDir, "/opt/code").WithEntrypoint("sudo").Run("chown", "-R", "opam", ".")
	if err != nil {
		log.WithError(err).Error("Error granting permissions to opam")
		return err
	}
	return nil
}
