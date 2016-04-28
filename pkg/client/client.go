package client

type client struct {
	unikIP string
}

func UnikClient(unikIP string) *client {
	return &client{unikIP: unikIP}
}

func (c *client) Images() *images {
	return &images{unikIP: c.unikIP}
}

func (c *client) Instances() *instances {
	return &instances{unikIP: c.unikIP}
}

func (c *client) Volumes() *volumes {
	return &volumes{unikIP: c.unikIP}
}