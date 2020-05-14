package deconz

import (
	"fmt"
	"net/http"
	"strings"
)

type Light struct {
	Id               string
	Hub				 *Deconz
	ETag             string                 `json:"etag"`
	HasColour        bool                   `json:"hascolor"`
	ManufacturerName string                 `json:"manufacturername"`
	ModelID          string                 `json:"modelid"`
	Name             string                 `json:"name"`
	State            map[string]interface{} `json:"state"`
	SoftwareVersion  string                 `json:"swversion"`
	Type             string                 `json:"type"`
	UniqueID         string                 `json:"uniqueid"`
}

func (l *Light) SetPower(on bool) error {
	rq, err := http.NewRequest("PUT",
		       fmt.Sprintf("http://%s/api/%s/lights/%s/state", l.Hub.HubAddress, l.Hub.Username, l.Id),
		       strings.NewReader(fmt.Sprintf(`{ "on": %t }`, on)))
	if err != nil {
		return fmt.Errorf("failed to make state change request: %w", err)
	}

	res,err := l.Hub.client.Do(rq)
	if err != nil {
		return fmt.Errorf("failed to make state change request: %w", err)
	}

	if res.StatusCode == 200 {
		l.State["on"] = on
		return nil
	}


	return fmt.Errorf("failed to set light state")
}

func (l *Light) SetBrightness(brightness int) error {
	rq, err := http.NewRequest("PUT",
		fmt.Sprintf("http://%s/api/%s/lights/%s/state", l.Hub.HubAddress, l.Hub.Username, l.Id),
		strings.NewReader(fmt.Sprintf(`{ "bri": %d }`, brightness)))
	if err != nil {
		return fmt.Errorf("failed to make state change request: %w", err)
	}

	res,err := l.Hub.client.Do(rq)
	if err != nil {
		return fmt.Errorf("failed to make state change request: %w", err)
	}

	if res.StatusCode == 200 {
		l.State["bri"] = brightness
		return nil
	}

	return fmt.Errorf("failed to set light state")
}