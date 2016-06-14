package qemu

func (p *QemuProvider) DeleteInstance(id string, force bool) error {
	return p.StopInstance(id)
}
