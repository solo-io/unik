package state

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"sync"
	"encoding/json"
	"os"
	"io/ioutil"
)

type LocalStorageState interface {
	State
	GetImagePaths() map[string]string
	GetVolumePaths() map[string]string
	ModifyImagePaths(modify func(imagePaths map[string]string) error) error
	ModifyVolumePaths(modify func(imagePaths map[string]string) error) error
}

type localStorageState struct {
	BaseState       State        `json:"BaseState"`
	ImagePaths      map[string]string `json:"ImagePaths"`
	VolumePaths     map[string]string `json:"VolumePaths"`
	saveLock        sync.Mutex
	imagePathsLock  sync.RWMutex
	volumePathsLock sync.RWMutex
	saveFile        string
}

func NewLocalStorageState(saveFile string) *localStorageState {
	return &localStorageState{
		BaseState: NewBasicState(""),
		ImagePaths: make(map[string]string),
		VolumePaths: make(map[string]string),
		saveFile:      saveFile,
	}
}

func LocalStorageStateFromFile(saveFile string) (*localStorageState, error) {
	data, err := ioutil.ReadFile(saveFile)
	if err != nil {
		return nil, lxerrors.New("error reading save file " + saveFile, err)
	}
	var s localStorageState
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, lxerrors.New("failed to unmarshal data " + string(data) + " to memory state", err)
	}
	s.saveFile = saveFile
	return &s, nil
}

func (s *localStorageState) GetImages() map[string]*types.Image {
	return s.BaseState.GetImages()
}

func (s *localStorageState) GetInstances() map[string]*types.Instance {
	return s.BaseState.GetInstances()
}

func (s *localStorageState) GetVolumes() map[string]*types.Volume {
	return s.BaseState.GetVolumes()
}

func (s *localStorageState) GetImagePaths() map[string]string {
	s.imagePathsLock.RLock()
	defer s.imagePathsLock.RUnlock()
	return s.ImagePaths
}

func (s *localStorageState) GetVolumePaths() map[string]string {
	s.volumePathsLock.RLock()
	defer s.volumePathsLock.RUnlock()
	return s.VolumePaths
}

func (s *localStorageState) ModifyImages(modify func(images map[string]*types.Image) error) error {
	return s.BaseState.ModifyImages(modify)
}

func (s *localStorageState) ModifyInstances(modify func(instances map[string]*types.Instance) error) error {
	return s.BaseState.ModifyInstances(modify)
}

func (s *localStorageState) ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error {
	return s.BaseState.ModifyVolumes(modify)
}

func (s *localStorageState) ModifyImagePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return modify(s.ImagePaths)
}

func (s *localStorageState) ModifyVolumePaths(modify func(imagePaths map[string]string) error) error {
	s.imagePathsLock.Lock()
	defer s.imagePathsLock.Unlock()
	return modify(s.VolumePaths)
}

func (s *localStorageState) Save() error {
	s.saveLock.Lock()
	defer s.saveLock.Unlock()
	data, err := json.Marshal(s)
	if err != nil {
		return lxerrors.New("failed to marshal memory state to json", err)
	}
	os.MkdirAll(filepath.Dir(s.saveFile), 0751)
	err = ioutil.WriteFile(s.saveFile, data, 0644)
	if err != nil {
		return lxerrors.New("writing save file " + s.saveFile, err)
	}
	return nil
}

