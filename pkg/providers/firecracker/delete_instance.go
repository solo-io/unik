package firecracker

func (p *FirecrackerProvider) DeleteInstance(id string, force bool) error {
	return p.StopInstance(id)
}
