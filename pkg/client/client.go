package client

import (
	"encoding/json"
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"net/http"
	"net/url"
	"strings"
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
	if err != nil {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var compilers []string
	if err := json.Unmarshal(body, &compilers); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return compilers, nil
}

func (c *client) AvailableProviders() ([]string, error) {
	resp, body, err := lxhttpclient.Get(c.unikIP, "/available_providers", nil)
	if err != nil {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var compilers []string
	if err := json.Unmarshal(body, &compilers); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return compilers, nil
}

func (c *client) DescribeCompiler(base string, lang string, provider string) (string, error) {
	query := buildQuery(map[string]interface{}{
		"base":     base,
		"lang":     lang,
		"provider": provider,
	})
	resp, body, err := lxhttpclient.Get(c.unikIP, "/describe_compiler"+query, nil)
	if err != nil {
		return "", errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return string(body), nil
}

func buildQuery(params map[string]interface{}) string {
	queryArray := []string{}
	for key, val := range params {
		queryArray = append(queryArray, url.QueryEscape(fmt.Sprintf("%s", key))+"="+url.QueryEscape(fmt.Sprintf("%v", val)))
	}
	queryString := "?" + strings.Join(queryArray, "&")
	return queryString
}
