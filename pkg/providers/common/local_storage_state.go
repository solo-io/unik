package common

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

type LocalStorageState struct {
	BaseState       state.State	`json:"BaseState"`
	ImagePaths      map[string]string `json:"ImagePaths"`
	VolumePaths     map[string]string `json:"VolumePaths"`
	saveLock        *sync.Mutex
	imagePathsLock  *sync.Mutex
	volumePathsLock *sync.Mutex
	saveFile        string
}

func NewLocalStorageState(saveFile string) *LocalStorageState {
	return &LocalStorageState{
		BaseState: state.NewMemoryState(""),
		ImagePaths: make(map[string]string),
		VolumePaths: make(map[string]string),
		saveLock:      &sync.Mutex{},
		imagePathsLock:      &sync.Mutex{},
		volumePathsLock:      &sync.Mutex{},
		saveFile:      saveFile,
	}
}

func (s *LocalStorageState) GetImages() map[string]*types.Image {
	return s.BaseState.GetImages()
}

func (s *LocalStorageState) GetInstances() map[string]*types.Instance {
	return s.BaseState.GetInstances()
}


func (s *LocalStorageState) GetVolumes() map[string]*types.Volume {
	return s.BaseState.GetVolumes()
}

func (s *LocalStorageState) GetImagePaths() map[string]string {
	return s.ImagePaths
}

func (s *LocalStorageState) GetVolumePaths() map[string]string {
	return s.VolumePaths
}

func (s *LocalStorageState) ModifyImages(modify func(images map[string]*types.Image) error) error {
	return s.BaseState.ModifyImages(modify)
}

func (s *LocalStorageState) ModifyInstances(modify func(instances map[string]*types.Instance) error) error {
	return s.BaseState.ModifyInstances(modify)
}

func (s *LocalStorageState) ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error {
	return s.BaseState.ModifyVolumes(modify)
}

func (s *LocalStorageState) ModifyImagePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return modify(s.ImagePaths)
}

func (s *LocalStorageState) ModifyVolumePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return modify(s.VolumePaths)
}

func (s *LocalStorageState) Save() error {
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

func (s *LocalStorageState) Load() error {
	data, err := ioutil.ReadFile(s.saveFile)
	if err != nil {
		return lxerrors.New("error reading save file "+s.saveFile, err)
	}
	var newS LocalStorageState
	err = json.Unmarshal(data, &newS)
	if err != nil {
		return lxerrors.New("failed to unmarshal data "+string(data)+" to memory state", err)
	}
	*s = newS
	return nil
}
