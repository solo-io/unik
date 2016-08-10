package qemu

func (p *XenProvider) DeleteInstance(id string, force bool) error {
	return p.StopInstance(id)
}
