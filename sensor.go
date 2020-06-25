package deconz

import "sync"

type Sensor struct {
	Id               string
	Hub				 *Deconz
	Config           map[string]interface{} `json:"config"`
	ETag             string                 `json:"etag"`
	ManufacturerName string                 `json:"manufacturername"`
	ModelID          string                 `json:"modelid"`
	Name             string                 `json:"name"`
	State            map[string]interface{} `json:"state"`
	SoftwareVersion  string                 `json:"swversion"`
	Type             string                 `json:"type"`
	UniqueID         string                 `json:"uniqueid"`
	m sync.Mutex
}

func (l *Sensor) updateState(state map[string]interface{}) {
	l.m.Lock()
	l.State = state
	l.m.Unlock()
}
