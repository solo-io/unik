package types

type RunInstanceParams struct {
	Name                 string
	ImageId              string
	MntPointsToVolumeIds map[string]string
	Env                  map[string]string
	InstanceMemory       int
	NoCleanup            bool
}

type StageImageParams struct {
	Name string
	RawImage *RawImage
	Force bool
	NoCleanup bool
}

type CreateVolumeParams struct {
	Name string
	ImagePath string
	NoCleanup bool
}

type CompileImageParams struct {
	SourcesDir string
	Args       string
	MntPoints  []string
}