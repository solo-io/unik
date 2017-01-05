package compilers

import (
	"fmt"
	"strings"

	"github.com/emc-advanced-dev/pkg/errors"
)

const (
	Rump = "rump"
)

type CompilerType string

func (c CompilerType) Base() string {
	return strings.Split(string(c), "-")[0]
}
func (c CompilerType) Language() string {
	return strings.Split(string(c), "-")[1]
}
func (c CompilerType) Provider() string {
	return strings.Split(string(c), "-")[2]
}
func (c CompilerType) String() string {
	return string(c)
}

var (
	//available compilers
	RUMP_C_XEN        = compilerName("rump", "c", "xen")
	RUMP_C_AWS        = compilerName("rump", "c", "aws")
	RUMP_C_VIRTUALBOX = compilerName("rump", "c", "virtualbox")
	RUMP_C_VSPHERE    = compilerName("rump", "c", "vsphere")
	RUMP_C_QEMU       = compilerName("rump", "c", "qemu")
	RUMP_C_PHOTON     = compilerName("rump", "c", "photon")
	RUMP_C_OPENSTACK  = compilerName("rump", "c", "openstack")

	RUMP_GO_XEN        = compilerName("rump", "go", "xen")
	RUMP_GO_AWS        = compilerName("rump", "go", "aws")
	RUMP_GO_VIRTUALBOX = compilerName("rump", "go", "virtualbox")
	RUMP_GO_VSPHERE    = compilerName("rump", "go", "vsphere")
	RUMP_GO_QEMU       = compilerName("rump", "go", "qemu")
	RUMP_GO_PHOTON     = compilerName("rump", "go", "photon")
	RUMP_GO_OPENSTACK  = compilerName("rump", "go", "openstack")
	RUMP_GO_GCLOUD     = compilerName("rump", "go", "gcloud")

	RUMP_NODEJS_XEN        = compilerName("rump", "nodejs", "xen")
	RUMP_NODEJS_AWS        = compilerName("rump", "nodejs", "aws")
	RUMP_NODEJS_VIRTUALBOX = compilerName("rump", "nodejs", "virtualbox")
	RUMP_NODEJS_VSPHERE    = compilerName("rump", "nodejs", "vsphere")
	RUMP_NODEJS_QEMU       = compilerName("rump", "nodejs", "qemu")
	RUMP_NODEJS_OPENSTACK  = compilerName("rump", "nodejs", "openstack")

	RUMP_PYTHON_XEN        = compilerName("rump", "python", "xen")
	RUMP_PYTHON_AWS        = compilerName("rump", "python", "aws")
	RUMP_PYTHON_VIRTUALBOX = compilerName("rump", "python", "virtualbox")
	RUMP_PYTHON_VSPHERE    = compilerName("rump", "python", "vsphere")
	RUMP_PYTHON_QEMU       = compilerName("rump", "python", "qemu")
	RUMP_PYTHON_OPENSTACK  = compilerName("rump", "python", "openstack")

	RUMP_JAVA_XEN        = compilerName("rump", "java", "xen")
	RUMP_JAVA_AWS        = compilerName("rump", "java", "aws")
	RUMP_JAVA_VIRTUALBOX = compilerName("rump", "java", "virtualbox")
	RUMP_JAVA_VSPHERE    = compilerName("rump", "java", "vsphere")
	RUMP_JAVA_QEMU       = compilerName("rump", "java", "qemu")
	RUMP_JAVA_OPENSTACK  = compilerName("rump", "java", "openstack")

	OSV_JAVA_XEN        = compilerName("osv", "java", "xen")
	OSV_JAVA_AWS        = compilerName("osv", "java", "aws")
	OSV_JAVA_VIRTUALBOX = compilerName("osv", "java", "virtualbox")
	OSV_JAVA_VSPHERE    = compilerName("osv", "java", "vsphere")
	OSV_JAVA_QEMU       = compilerName("osv", "java", "qemu")
	OSV_JAVA_OPENSTACK  = compilerName("osv", "java", "openstack")

	OSV_NODEJS_QEMU      = compilerName("osv", "nodejs", "qemu")
	OSV_NODEJS_OPENSTACK = compilerName("osv", "nodejs", "openstack")

	OSV_NATIVE_QEMU      = compilerName("osv", "native", "qemu")
	OSV_NATIVE_OPENSTACK = compilerName("osv", "native", "openstack")

	INCLUDEOS_CPP_QEMU       = compilerName("includeos", "cpp", "qemu")
	INCLUDEOS_CPP_XEN        = compilerName("includeos", "cpp", "xen")
	INCLUDEOS_CPP_VIRTUALBOX = compilerName("includeos", "cpp", "virtualbox")
	INCLUDEOS_CPP_OPENSTACK  = compilerName("includeos", "cpp", "openstack")

	MIRAGE_OCAML_XEN  = compilerName("mirage", "ocaml", "xen")
	MIRAGE_OCAML_UKVM = compilerName("mirage", "ocaml", "ukvm")
	MIRAGE_OCAML_QEMU = compilerName("mirage", "ocaml", "qemu")
)

var compilers = []CompilerType{
	RUMP_C_XEN,
	RUMP_C_AWS,
	RUMP_C_VIRTUALBOX,
	RUMP_C_VSPHERE,
	RUMP_C_QEMU,
	RUMP_C_PHOTON,
	RUMP_C_OPENSTACK,

	RUMP_GO_XEN,
	RUMP_GO_AWS,
	RUMP_GO_VIRTUALBOX,
	RUMP_GO_VSPHERE,
	RUMP_GO_QEMU,
	RUMP_GO_PHOTON,
	RUMP_GO_OPENSTACK,
	RUMP_GO_GCLOUD,

	RUMP_NODEJS_XEN,
	RUMP_NODEJS_AWS,
	RUMP_NODEJS_VIRTUALBOX,
	RUMP_NODEJS_VSPHERE,
	RUMP_NODEJS_QEMU,
	RUMP_NODEJS_OPENSTACK,

	RUMP_PYTHON_XEN,
	RUMP_PYTHON_AWS,
	RUMP_PYTHON_VIRTUALBOX,
	RUMP_PYTHON_VSPHERE,
	RUMP_PYTHON_QEMU,
	RUMP_PYTHON_OPENSTACK,

	RUMP_JAVA_XEN,
	RUMP_JAVA_AWS,
	RUMP_JAVA_VIRTUALBOX,
	RUMP_JAVA_VSPHERE,
	RUMP_JAVA_QEMU,
	RUMP_JAVA_OPENSTACK,

	OSV_JAVA_XEN,
	OSV_JAVA_AWS,
	OSV_JAVA_VIRTUALBOX,
	OSV_JAVA_VSPHERE,
	OSV_JAVA_QEMU,
	OSV_JAVA_OPENSTACK,

	OSV_NODEJS_QEMU,
	OSV_NODEJS_OPENSTACK,

	OSV_NATIVE_QEMU,
	OSV_NATIVE_OPENSTACK,

	INCLUDEOS_CPP_QEMU,
	INCLUDEOS_CPP_XEN,
	INCLUDEOS_CPP_VIRTUALBOX,
	INCLUDEOS_CPP_OPENSTACK,

	MIRAGE_OCAML_XEN,
	MIRAGE_OCAML_UKVM,
	MIRAGE_OCAML_QEMU,
}

func ValidateCompiler(base, language, provider string) (CompilerType, error) {
	baseMatch := false
	languageMatch := false
	for _, compiler := range compilers {
		if compiler.Base() == base {
			baseMatch = true
		}
		if compiler.Base() == base && compiler.Language() == language {
			languageMatch = true
		}
		if compiler.Base() == base && compiler.Language() == language && compiler.Provider() == provider {
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
		uniqueBases[compiler.Base()] = struct{}{}
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
		if compiler.Base() == base {
			uniqueLangs[compiler.Language()] = struct{}{}
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
		if (compiler.Base() == base) && (compiler.Language() == language) {
			uniqueProviders[compiler.Provider()] = struct{}{}
		}
	}
	providers := []string{}
	for provider := range uniqueProviders {
		providers = append(providers, provider)
	}
	return providers
}

func compilerName(base, language, provider string) CompilerType {
	return CompilerType(fmt.Sprintf("%s-%s-%s", base, language, provider))
}
