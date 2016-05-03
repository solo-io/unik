package client

import (
	"fmt"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"net/http"
	"github.com/emc-advanced-dev/pkg/errors"
	"encoding/json"
	"strings"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io"
)

type instances struct {
	unikIP string
}

const envDelimiter = "DEFAULT_DELIMETER"
const envPairDelimiter = "DEFAULT_PAIR_DELIMETER"
const mntDelimiter = ","
const mntPairDelimiter = ":"

func (i *instances) All() ([]*types.Instance, error) {
	resp, body, err := lxhttpclient.Get(i.unikIP, "/instances", nil)
	if err != nil  {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var instances []*types.Instance
	if err := json.Unmarshal(body, &instances); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type []*types.Instance", string(body)), err)
	}
	return instances, nil
}

func (i *instances) Get(id string) (*types.Instance, error) {
	resp, body, err := lxhttpclient.Get(i.unikIP, "/instances/"+id, nil)
	if err != nil  {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), nil)
	}
	var instance types.Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Instance", string(body)), err)
	}
	return &instance, nil
}

func (i *instances) Delete(id string, force bool) error {
	query := fmt.Sprintf("?force=%v", force)
	resp, body, err := lxhttpclient.Delete(i.unikIP, "/instances/"+id+query, nil)
	if err != nil  {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}

func (i *instances) GetLogs(id string) (string, error) {
	resp, body, err := lxhttpclient.Get(i.unikIP, "/instances/"+id+"/logs", nil)
	if err != nil  {
		return "", errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return string(body), nil
}

func (i *instances) AttachLogs(id string, deleteOnDisconnect bool) (io.ReadCloser, error) {
	query := fmt.Sprintf("?follow=%v&delete=%v", true, deleteOnDisconnect)
	resp, err := lxhttpclient.GetAsync(i.unikIP, "/instances/"+id+"/logs"+query, nil)
	if err != nil  {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed with status %v", resp.StatusCode), err)
	}
	return resp.Body, nil
}

func (i *instances) Run(instanceName, imageName string, mounts, env map[string]string, noCleanup bool) (*types.Instance, error) {
	envPairs := []string{}
	for key, val := range env {
		envPairs = append(envPairs, fmt.Sprintf("%s%s%s", key, envPairDelimiter, val))
	}
	envStr := strings.Join(envPairs, envDelimiter)

	mntPairs := []string{}
	for key, val := range mounts {
		mntPairs = append(mntPairs, fmt.Sprintf("%s%s%s", key, mntPairDelimiter, val))
	}
	mntStr := strings.Join(mntPairs, mntDelimiter)

	query := fmt.Sprintf("?image_name=%s&useDelimiter=%s&usePairDelimiter=%s&env=%s&mounts=%s&no_cleanup=%v", imageName, envDelimiter, envPairDelimiter, envStr, mntStr, noCleanup)
	resp, body, err := lxhttpclient.Post(i.unikIP, "/instances/"+instanceName+"/run"+query, nil, nil)
	if err != nil  {
		return nil, errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	var instance types.Instance
	if err := json.Unmarshal(body, &instance); err != nil {
		return nil, errors.New(fmt.Sprintf("response body %s did not unmarshal to type *types.Instance", string(body)), err)
	}
	return &instance, nil
}

func (i *instances) Start(id string) error {
	resp, body, err := lxhttpclient.Post(i.unikIP, "/instances/"+id+"/start", nil, nil)
	if err != nil  {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}

func (i *instances) Stop(id string) error {
	resp, body, err := lxhttpclient.Post(i.unikIP, "/instances/"+id+"/stop", nil, nil)
	if err != nil  {
		return errors.New("request failed", err)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)), err)
	}
	return nil
}