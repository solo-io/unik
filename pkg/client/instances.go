package client

import (
	"fmt"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"net/http"
	"github.com/emc-advanced-dev/pkg/errors"
	"encoding/json"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
)

type instances struct {
	unikIP string
}

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

func (i *instances) Run(instanceName, imageName string, mountPointsToVols, env map[string]string, memoryMb int, noCleanup, debugMode bool) (*types.Instance, error) {
	runInstanceRequest := daemon.RunInstanceRequest{
		InstanceName: instanceName,
		ImageName: imageName,
		Mounts: mountPointsToVols,
		Env: env,
		MemoryMb: memoryMb,
		NoCleanup: noCleanup,
		DebugMode: debugMode,
	}
	resp, body, err := lxhttpclient.Post(i.unikIP, "/instances/run", nil, runInstanceRequest)
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