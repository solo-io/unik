package vsphere

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"sync"
	"encoding/json"
	"os"
	"io/ioutil"
)

type vsphereState struct {
	BaseState       state.State	`json:"BaseState"`
	ImagePaths      map[string]string `json:"ImagePaths"`
	VolumePaths     map[string]string `json:"VolumePaths"`
	saveLock        *sync.Mutex
	imagePathsLock  *sync.Mutex
	volumePathsLock *sync.Mutex
	saveFile        string
}

func newVsphereState(saveFile string) *vsphereState {
	return &vsphereState{
		BaseState: state.NewMemoryState(""),
		ImagePaths: make(map[string]string),
		VolumePaths: make(map[string]string),
		saveLock:      &sync.Mutex{},
		imagePathsLock:      &sync.Mutex{},
		volumePathsLock:      &sync.Mutex{},
		saveFile:      saveFile,
	}
}

func (s *vsphereState) GetImages() map[string]*types.Image {
	return s.BaseState.GetImages()
}

func (s *vsphereState) GetInstances() map[string]*types.Instance {
	return s.BaseState.GetInstances()
}


func (s *vsphereState) GetVolumes() map[string]*types.Volume {
	return s.BaseState.GetVolumes()
}

func (s *vsphereState) GetImagePaths() map[string]string {
	return s.ImagePaths
}

func (s *vsphereState) GetVolumePaths() map[string]string {
	return s.VolumePaths
}

func (s *vsphereState) ModifyImages(modify func(images map[string]*types.Image) error) error {
	return s.BaseState.ModifyImages(modify)
}

func (s *vsphereState) ModifyInstances(modify func(instances map[string]*types.Instance) error) error {
	return s.BaseState.ModifyInstances(modify)
}

func (s *vsphereState) ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error {
	return s.BaseState.ModifyVolumes(modify)
}

func (s *vsphereState) ModifyImagePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return s.BaseState.ModifyInstances(modify)
}

func (s *vsphereState) ModifyVolumePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return s.BaseState.ModifyInstances(modify)
}

func (s *vsphereState) Save() error {
	s.saveLock.Lock()
	defer s.saveLock.Unlock()
	data, err := json.Marshal(s)
	if err != nil {
		return lxerrors.New("failed to marshal memory state to json", err)
	}
	os.MkdirAll(filepath.Dir(s.saveFile), 0644)
	err = ioutil.WriteFile(s.saveFile, data, 0644)
	if err != nil {
		return lxerrors.New("writing save file "+s.saveFile, err)
	}
	return nil
}

func (s *vsphereState) Load() error {
	data, err := ioutil.ReadFile(s.saveFile)
	if err != nil {
		return lxerrors.New("error reading save file "+s.saveFile, err)
	}
	var newS vsphereState
	err = json.Unmarshal(data, &newS)
	if err != nil {
		return lxerrors.New("failed to unmarshal data "+string(data)+" to memory state", err)
	}
	*s = newS
	return nil
}
