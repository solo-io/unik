package compilers

import (
	"fmt"
	"github.com/solo-io/unik/pkg/types"
	"strings"
)

type Compiler interface {
	CompileRawImage(params types.CompileImageParams) (*types.RawImage, error)

	// Usage describes how to prepare project to run it with UniK
	// The returned text should describe what configuration files to
	// prepare and how.
	Usage() *CompilerUsage
}

type CompilerUsage struct {
	// PrepareApplication section briefly describes how user should
	// prepare her application PRIOR composing unikernel with UniK
	PrepareApplication string

	// ConfigurationFiles lists configuration files needed by UniK.
	// It is a map filename:content_description.
	ConfigurationFiles map[string]string

	// Other is arbitrary content that will be printed at the end.
	Other string
}

func (c *CompilerUsage) ToString() string {
	prepApp := strings.TrimSpace(c.PrepareApplication)
	other := strings.TrimSpace(c.Other)

	configFiles := ""
	for k := range c.ConfigurationFiles {
		configFiles += fmt.Sprintf("------ %s ------\n", k)
		configFiles += strings.TrimSpace(c.ConfigurationFiles[k])
		configFiles += "\n\n"
	}
	configFiles = strings.TrimSpace(configFiles)

	description := fmt.Sprintf(`
HOW TO PREPARE APPLICATION	
%s

CONFIGURATION FILES
%s
`, prepApp, configFiles)

	if other != "" {
		description += fmt.Sprintf("\nOTHER\n%s", other)
	}

	return description
}
