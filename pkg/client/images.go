package client

import (
	"encoding/json"
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/config"
	"github.com/solo-io/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"net/http"
	"strings"
)

type images struct {
	unikIP string
}

func (i *images) All() ([]*types.Image, error) {
	resp, body, err := lxhttpclient.Get(i.unikIP, "/images", nil)
	if err != nil {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var images []*types.Image
	if err := json.Unmarshal(body, &images); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type []*types.Image", string(body)), err)
	}
	return images, nil
}

func (i *images) Get(id string) (*types.Image, error) {
	resp, body, err := lxhttpclient.Get(i.unikIP, "/images/"+id, nil)
	if err != nil {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var image types.Image
	if err := json.Unmarshal(body, &image); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return &image, nil
}

func (i *images) Build(name, sourceTar, base, lang, provider, args string, mounts []string, force, noCleanup bool) (*types.Image, error) {
	query := buildQuery(map[string]interface{}{
		"base":       base,
		"lang":       lang,
		"provider":   provider,
		"args":       args,
		"mounts":     strings.Join(mounts, ","),
		"force":      force,
		"no_cleanup": noCleanup,
	})
	resp, body, err := lxhttpclient.PostFile(i.unikIP, "/images/"+name+"/create"+query, "tarfile", sourceTar)
	if err != nil {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	var image types.Image
	if err := json.Unmarshal(body, &image); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Image", string(body)), err)
	}
	return &image, nil
}

func (i *images) Delete(id string, force bool) error {
	query := buildQuery(map[string]interface{}{
		"force": force,
	})
	resp, body, err := lxhttpclient.Delete(i.unikIP, "/images/"+id+query, nil)
	if err != nil {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}

func (i *images) Push(c config.HubConfig, imageName string) error {
	resp, body, err := lxhttpclient.Post(i.unikIP, "/images/push/"+imageName, nil, c)
	if err != nil {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}

func (i *images) Pull(c config.HubConfig, imageName, provider string, force bool) error {
	query := buildQuery(map[string]interface{}{
		"provider": provider,
		"force":    force,
	})
	resp, body, err := lxhttpclient.Post(i.unikIP, "/images/pull/"+imageName+query, nil, c)
	if err != nil {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}

func (i *images) RemoteDelete(c config.HubConfig, imageName string) error {
	resp, body, err := lxhttpclient.Post(i.unikIP, "/images/remote-delete/"+imageName, nil, c)
	if err != nil {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}
