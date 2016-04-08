package state

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"sync"
)

type State struct {
	lock      *sync.Mutex
	Instances map[string]*types.Instance `json:"Instances"`
	Images    map[string]*types.Image    `json:"Images"`
	Volumes   map[string]*types.Volume   `json:"Volumes"`
}
