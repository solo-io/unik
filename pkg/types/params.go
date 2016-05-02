package types

type RunInstanceParams struct {
	Name string
	ImageId string
	MntPointsToVolumeIds map[string]string
	Env map[string]string
	NoCleanup bool
}

type StageImageParams struct {
	Name string
	RawImage *RawImage
	Force bool
}

type CreateVolumeParams struct {
	Name string
	ImagePath string
}