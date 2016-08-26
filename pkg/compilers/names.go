package compilers

import (
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"strings"
)

const (
	Rump = "rump"
)

var (
	//available compilers
	RUMP_C_XEN        = compilerName("rump", "c", "xen")
	RUMP_C_AWS        = compilerName("rump", "c", "aws")
	RUMP_C_VIRTUALBOX = compilerName("rump", "c", "virtualbox")
	RUMP_C_VMWARE     = compilerName("rump", "c", "vmware")
	RUMP_C_QEMU       = compilerName("rump", "c", "qemu")
	RUMP_C_PHOTON     = compilerName("rump", "c", "photon")

	RUMP_GO_XEN        = compilerName("rump", "go", "xen")
	RUMP_GO_AWS        = compilerName("rump", "go", "aws")
	RUMP_GO_VIRTUALBOX = compilerName("rump", "go", "virtualbox")
	RUMP_GO_VMWARE     = compilerName("rump", "go", "vmware")
	RUMP_GO_QEMU       = compilerName("rump", "go", "qemu")
	RUMP_GO_PHOTON     = compilerName("rump", "go", "photon")

	RUMP_NODEJS_XEN        = compilerName("rump", "nodejs", "xen")
	RUMP_NODEJS_AWS        = compilerName("rump", "nodejs", "aws")
	RUMP_NODEJS_VIRTUALBOX = compilerName("rump", "nodejs", "virtualbox")
	RUMP_NODEJS_VMWARE     = compilerName("rump", "nodejs", "vmware")
	RUMP_NODEJS_QEMU       = compilerName("rump", "nodejs", "qemu")

	RUMP_PYTHON_XEN        = compilerName("rump", "python", "xen")
	RUMP_PYTHON_AWS        = compilerName("rump", "python", "aws")
	RUMP_PYTHON_VIRTUALBOX = compilerName("rump", "python", "virtualbox")
	RUMP_PYTHON_VMWARE     = compilerName("rump", "python", "vmware")
	RUMP_PYTHON_QEMU       = compilerName("rump", "python", "qemu")

	RUMP_JAVA_XEN        = compilerName("rump", "java", "xen")
	RUMP_JAVA_AWS        = compilerName("rump", "java", "aws")
	RUMP_JAVA_VIRTUALBOX = compilerName("rump", "java", "virtualbox")
	RUMP_JAVA_VMWARE     = compilerName("rump", "java", "vmware")
	RUMP_JAVA_QEMU       = compilerName("rump", "java", "qemu")

	OSV_JAVA_XEN        = compilerName("osv", "java", "xen")
	OSV_JAVA_AWS        = compilerName("osv", "java", "aws")
	OSV_JAVA_VIRTUALBOX = compilerName("osv", "java", "virtualbox")
	OSV_JAVA_VMAWRE     = compilerName("osv", "java", "vmware")
	OSV_JAVA_QEMU       = compilerName("osv", "java", "qemu")
	OSV_JAVA_OPENSTACK  = compilerName("osv", "java", "openstack")

	INCLUDEOS_CPP_QEMU       = compilerName("includeos", "cpp", "qemu")
	INCLUDEOS_CPP_XEN        = compilerName("includeos", "cpp", "xen")
	INCLUDEOS_CPP_VIRTUALBOX = compilerName("includeos", "cpp", "virtualbox")
)

var compilers = []string{
	RUMP_C_XEN,
	RUMP_C_AWS,
	RUMP_C_VIRTUALBOX,
	RUMP_C_VMWARE,
	RUMP_C_QEMU,
	RUMP_C_PHOTON,

	RUMP_GO_XEN,
	RUMP_GO_AWS,
	RUMP_GO_VIRTUALBOX,
	RUMP_GO_VMWARE,
	RUMP_GO_QEMU,
	RUMP_GO_PHOTON,

	RUMP_NODEJS_XEN,
	RUMP_NODEJS_AWS,
	RUMP_NODEJS_VIRTUALBOX,
	RUMP_NODEJS_VMWARE,
	RUMP_NODEJS_QEMU,

	RUMP_PYTHON_XEN,
	RUMP_PYTHON_AWS,
	RUMP_PYTHON_VIRTUALBOX,
	RUMP_PYTHON_VMWARE,
	RUMP_PYTHON_QEMU,

	RUMP_JAVA_XEN,
	RUMP_JAVA_AWS,
	RUMP_JAVA_VIRTUALBOX,
	RUMP_JAVA_VMWARE,
	RUMP_JAVA_QEMU,

	OSV_JAVA_XEN,
	OSV_JAVA_AWS,
	OSV_JAVA_VIRTUALBOX,
	OSV_JAVA_VMAWRE,
	OSV_JAVA_QEMU,
	OSV_JAVA_OPENSTACK,

	INCLUDEOS_CPP_QEMU,
	INCLUDEOS_CPP_XEN,
	INCLUDEOS_CPP_VIRTUALBOX,
}

func ValidateCompiler(base, language, provider string) (string, error) {
	baseMatch := false
	languageMatch := false
	for _, compiler := range compilers {
		if strings.HasPrefix(compiler, base) {
			baseMatch = true
		}
		if strings.HasPrefix(compiler, base+"-"+language) {
			baseMatch = true
		}
		if compiler == compilerName(base, language, provider) {
			return compiler, nil
		}
	}
	if !baseMatch {
		return "", errors.New("no compiler found for base "+base+", available bases: "+strings.Join(bases(), " | "), nil)
	}
	if !languageMatch {
		return "", errors.New("language "+language+" not found for base "+base+", available languages: "+strings.Join(langsForBase(base), " | "), nil)
	}
	return "", errors.New("provider "+provider+" not supported for unikernel runtime "+base+"-"+language+", available providers: "+strings.Join(providersForBaseLang(base, language), " | "), nil)
}

func bases() []string {
	uniqueBases := make(map[string]interface{})
	for _, compiler := range compilers {
		uniqueBases[strings.Split(compiler, "-")[0]] = struct{}{}
	}
	bases := []string{}
	for base := range uniqueBases {
		bases = append(bases, base)
	}
	return bases
}

func langsForBase(base string) []string {
	uniqueLangs := make(map[string]interface{})
	for _, compiler := range compilers {
		if strings.HasPrefix(compiler, base) {
			uniqueLangs[strings.Split(compiler, "-")[1]] = struct{}{}
		}
	}
	langs := []string{}
	for lang := range uniqueLangs {
		langs = append(langs, lang)
	}
	return langs
}

func providersForBaseLang(base, language string) []string {
	uniqueProviders := make(map[string]interface{})
	for _, compiler := range compilers {
		if strings.HasPrefix(compiler, base+"-"+language) {
			uniqueProviders[strings.Split(compiler, "-")[2]] = struct{}{}
		}
	}
	providers := []string{}
	for provider := range uniqueProviders {
		providers = append(providers, provider)
	}
	return providers
}

func compilerName(base, language, provider string) string {
	return fmt.Sprintf("%s-%s-%s", base, language, provider)
}
