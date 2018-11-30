package mirage

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"

	"github.com/solo-io/unik/pkg/compilers"
	unikos "github.com/solo-io/unik/pkg/os"
	"github.com/solo-io/unik/pkg/types"
	unikutil "github.com/solo-io/unik/pkg/util"
	"gopkg.in/yaml.v2"
)

type Type int

const (
	XenType Type = iota
	UKVMType
	VirtioType
)

type MirageCompiler struct {
	Type Type
}

func (c *MirageCompiler) CompileRawImage(params types.CompileImageParams) (*types.RawImage, error) {

	sourcesDir := params.SourcesDir

	if err := grantOpamPermissions(sourcesDir); err != nil {
		return nil, err
	}
	var containerToUse string
	var args []string
	switch c.Type {
	case XenType:
		var err error
		args, err = parseMirageManifest(sourcesDir)
		if err != nil {
			args, err = introspectArguments(sourcesDir)
			if err != nil {
				return nil, err
			}
		}
		containerToUse = "compilers-mirage-ocaml-xen"
		args = append([]string{"configure", "-t", "xen"}, args...)

	case VirtioType:
		containerToUse = "compilers-mirage-ocaml-ukvm"
		args = append([]string{"configure", "-t", "virtio"}, args...)

	case UKVMType:
		containerToUse = "compilers-mirage-ocaml-ukvm"
		args = append([]string{"configure", "-t", "ukvm"}, args...)
	default:
		return nil, errors.New("unknown type", nil)
	}

	if err := unikutil.NewContainer(containerToUse).WithEntrypoint("mirage").WithVolume(sourcesDir, "/opt/code").Run(args...); err != nil {
		return nil, err
	}

	if err := unikutil.NewContainer(containerToUse).WithEntrypoint("/usr/bin/make").WithVolume(sourcesDir, "/opt/code").Run(); err != nil {
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
	switch c.Type {
	case XenType:
		return c.packageForXen(sourcesDir, disks, params.NoCleanup)
	case UKVMType:
		return c.packageForUkvm(sourcesDir, disks, params.NoCleanup)
	case VirtioType:
		return c.packageForVirtio(sourcesDir, disks, params.NoCleanup)
	default:
		return nil, errors.New("unknown type", nil)
	}
}

func (c *MirageCompiler) packageForXen(sourcesDir string, disks []string, cleanup bool) (*types.RawImage, error) {
	// TODO: ukvm package zipfile for ukvm
	var res types.RawImage
	res.RunSpec.Compiler = compilers.MIRAGE_OCAML_XEN.String()

	unikernelfile, err := getUnikernelFile(sourcesDir)
	if err != nil {
		return nil, err
	}

	res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "/", DeviceName: "/dev/sda1"})
	for _, disk := range disks {
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: "xen:" + disk, DeviceName: disk})
	}

	// TODO: ukvm package zipfile for ukvm
	imgFile, err := compilers.BuildBootableImage(unikernelfile, "", false, cleanup)

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

func (c *MirageCompiler) packageForUkvm(sourcesDir string, disks []string, cleanup bool) (*types.RawImage, error) {
	return c.packageUnikernel(sourcesDir, disks, cleanup, "ukvm")
}
func (c *MirageCompiler) packageForVirtio(sourcesDir string, disks []string, cleanup bool) (*types.RawImage, error) {
	// for qemu we create an empty cmdline file
	r, err := c.packageUnikernel(sourcesDir, disks, cleanup, "virtio")
	if err != nil {
		return r, err
	}
	return r, err
}

func (c *MirageCompiler) packageUnikernel(sourcesDir string, disks []string, cleanup bool, unikernel string) (*types.RawImage, error) {
	// find ukvm-bin -> the monitor
	// find *.ukvm -> the unikernel

	matches, err := filepath.Glob(filepath.Join(sourcesDir, "*."+unikernel))
	if err != nil {
		return nil, err
	}

	// filter non relevant Makefile.ukvm
	var potentialMatches []string
	for _, m := range matches {
		if !strings.HasSuffix(m, "Makefile."+unikernel) {
			potentialMatches = append(potentialMatches, m)
		}
	}

	if len(potentialMatches) != 1 {
		return nil, errors.New(fmt.Sprintf("Ukvm kernel file count is wrong: %v", potentialMatches), nil)
	}

	kernel := potentialMatches[0]

	// place them in the image directory

	tmpImageDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	if err := unikos.CopyFile(kernel, filepath.Join(tmpImageDir, "program.bin")); err != nil {
		return nil, errors.New("copying bootable image to image dir", err)
	}

	if unikernel == "ukvm" {
		monitor := filepath.Join(sourcesDir, "ukvm-bin")
		if err := unikos.CopyFile(monitor, filepath.Join(tmpImageDir, "ukvm-bin")); err != nil {
			return nil, errors.New("copying bootable image to image dir", err)
		}
	} else {
		// it is virtio for qemu
		// mirage doesn't have command line, so create an empty one
		// this is the hint for qemu to use PV mode
		f, err := os.Create(filepath.Join(tmpImageDir, "cmdline"))
		if err != nil {
			return nil, errors.New("creating empty cmdline for image", err)
		}
		f.Close()
	}

	res := &types.RawImage{}
	for _, disk := range disks {
		res.RunSpec.DeviceMappings = append(res.RunSpec.DeviceMappings, types.DeviceMapping{MountPoint: unikernel + ":" + disk, DeviceName: disk})
	}
	if unikernel == "ukvm" {
		res.RunSpec.Compiler = compilers.MIRAGE_OCAML_UKVM.String()
	} else {
		res.RunSpec.Compiler = compilers.MIRAGE_OCAML_QEMU.String()
	}
	res.LocalImagePath = tmpImageDir
	res.StageSpec = types.StageSpec{
		ImageFormat: types.ImageFormat_Folder,
	}
	res.RunSpec.DefaultInstanceMemory = 256

	return res, nil
}

func (r *MirageCompiler) Usage() *compilers.CompilerUsage {
	return nil
}

var parseRegEx = regexp.MustCompile(`vdev=(\S+),\s.+?target=@\S+?:(\S+?)@`)

func getUnikernelFile(sourcesDir string) (string, error) {

	matches, err := filepath.Glob(filepath.Join(sourcesDir, "*.xen"))
	if err != nil {
		return "", err
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
