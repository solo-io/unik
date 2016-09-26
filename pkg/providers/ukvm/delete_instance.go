package ukvm

func (p *UkvmProvider) DeleteInstance(id string, force bool) error {
	return p.StopInstance(id)
}
