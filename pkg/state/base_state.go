package state

import (
	"encoding/json"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type basicState struct {
	imagesLock    sync.RWMutex
	instancesLock sync.RWMutex
	volumesLock   sync.RWMutex
	saveLock      sync.Mutex
	saveFile      string
	Images        map[string]*types.Image    `json:"Images"`
	Instances     map[string]*types.Instance `json:"Instances"`
	Volumes       map[string]*types.Volume   `json:"Volumes"`
}

func NewBasicState(saveFile string) *basicState {
	return &basicState{
		saveFile:  saveFile,
		Images:    make(map[string]*types.Image),
		Instances: make(map[string]*types.Instance),
		Volumes:   make(map[string]*types.Volume),
	}
}

func BasicStateFromFile(saveFile string) (*basicState, error) {
	data, err := ioutil.ReadFile(saveFile)
	if err != nil {
		return nil, errors.New("error reading save file "+saveFile, err)
	}
	var s basicState
	err = json.Unmarshal(data, &s)
	if err != nil {
		return nil, errors.New("failed to unmarshal data "+string(data)+" to memory state", err)
	}
	s.saveFile = saveFile
	return &s, nil
}

func (s *basicState) GetImages() map[string]*types.Image {
	s.imagesLock.RLock()
	defer s.imagesLock.RUnlock()
	imagesCopy := make(map[string]*types.Image)
	for id, image := range s.Images {
		imageCopy := image.Copy()
		imagesCopy[id] = imageCopy
	}
	return imagesCopy
}

func (s *basicState) GetInstances() map[string]*types.Instance {
	s.instancesLock.RLock()
	defer s.instancesLock.RUnlock()
	instancesCopy := make(map[string]*types.Instance)
	for id, instance := range s.Instances {
		instanceCopy := instance.Copy()
		instancesCopy[id] = instanceCopy
	}
	return instancesCopy
}

func (s *basicState) GetVolumes() map[string]*types.Volume {
	s.volumesLock.RLock()
	defer s.volumesLock.RUnlock()
	volumesCopy := make(map[string]*types.Volume)
	for id, volume := range s.Volumes {
		volumeCopy := volume.Copy()
		volumesCopy[id] = volumeCopy
	}
	return volumesCopy
}

func (s *basicState) ModifyImages(modify func(images map[string]*types.Image) error) error {
	s.imagesLock.Lock()
	defer s.imagesLock.Unlock()
	return modify(s.Images)
}

func (s *basicState) ModifyInstances(modify func(instances map[string]*types.Instance) error) error {
	s.instancesLock.Lock()
	defer s.instancesLock.Unlock()
	return modify(s.Instances)
}

func (s *basicState) ModifyVolumes(modify func(volumes map[string]*types.Volume) error) error {
	s.volumesLock.Lock()
	defer s.volumesLock.Unlock()
	return modify(s.Volumes)
}

func (s *basicState) Save() error {
	s.saveLock.Lock()
	defer s.saveLock.Unlock()
	data, err := json.Marshal(s)
	if err != nil {
		return errors.New("failed to marshal memory state to json", err)
	}
	os.MkdirAll(filepath.Dir(s.saveFile), 0755)
	err = ioutil.WriteFile(s.saveFile, data, 0644)
	if err != nil {
		return errors.New("writing save file "+s.saveFile, err)
	}
	return nil
}
