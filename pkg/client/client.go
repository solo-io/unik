package client

import (
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/layer-x/layerx-commons/lxerrors"
	"fmt"
	"net/http"
	"encoding/json"
)

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

func (c *client) AvailableCompilers() ([]string, error) {
	resp, body, err := lxhttpclient.Get(c.unikIP, "/available_compilers", nil)
	if err != nil  {
		return nil, lxerrors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, lxerrors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var compilers []string
	if err := json.Unmarshal(body, &compilers); err != nil {
		return nil, lxerrors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return compilers, nil
}

func (c *client) AvailableProviders() ([]string, error) {
	resp, body, err := lxhttpclient.Get(c.unikIP, "/available_providers", nil)
	if err != nil  {
		return nil, lxerrors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, lxerrors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var compilers []string
	if err := json.Unmarshal(body, &compilers); err != nil {
		return nil, lxerrors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return compilers, nil
}