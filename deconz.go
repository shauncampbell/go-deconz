package deconz

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"time"
)

type SensorFoundFunc func ( d *Deconz, uniqueId string, sensor *Sensor )
type LightFoundFunc func (  d *Deconz, uniqueId string, light *Light )
type SensorStateChangeFunc func (  d *Deconz, uniqueId string, state map[string]interface{})

type Deconz struct {
	HubAddress string
	Username string
	sensors map[string]*Sensor
	lights map[string]*Light
	client *http.Client
	OnSensorFound SensorFoundFunc
	OnLightFound LightFoundFunc
	OnSensorStateChange SensorStateChangeFunc
	OnLightStateChange SensorStateChangeFunc
}

type Event struct {
	Event string `json:"e"`
	Id string `json:"id"`
	R string `json:"r"`
	State map[string]interface{} `json:"state"`
	T string `json:"t"`
	UniqueID string `json:"uniqueid"`
}

func NewDeconz(hubAddress, username string) (*Deconz, error) {
	d := Deconz{
		HubAddress: hubAddress,
		Username: username,
		sensors: make(map[string]*Sensor),
		lights: make(map[string]*Light),
	}
	d.client = http.DefaultClient
	return &d, nil
}

func (d *Deconz) Lights() ([]*Light, error) {
	lights := make([]*Light, 0)

	if err := d.updateLights(); err != nil {
		return nil, err
	}

	for _, v := range d.lights {
		lights = append(lights, v)
	}

	return lights, nil
}


func (d *Deconz) Sensors() ([]*Sensor, error) {
	sensors := make([]*Sensor, 0)

	if err := d.updateSensors(); err != nil {
		return nil, err
	}

	for _, v := range d.sensors {
		sensors = append(sensors, v)
	}

	return sensors, nil
}

func (d *Deconz) Scan() error {
	if err := d.updateSensors(); err != nil {
		return fmt.Errorf("failed to update sensor list: %w", err)
	}

	if err := d.updateLights(); err != nil {
		return fmt.Errorf("failed to update lights list: %w", err)
	}
	u := url.URL{Scheme: "ws", Host: d.HubAddress + ":443", Path: ""}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	//defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				time.Sleep(30 * time.Second)
				continue
			}
			var event Event
			if err = json.Unmarshal(message, &event); err != nil {
				fmt.Printf("err: %s", err)
			}

			if d.sensors[event.UniqueID] != nil && d.OnSensorStateChange != nil {
				sensor := d.sensors[event.UniqueID]
				sensor.updateState(event.State)
				go d.OnSensorStateChange(d, event.UniqueID, event.State)
			} else if d.lights[event.UniqueID] != nil && d.OnLightStateChange != nil {
				light := d.lights[event.UniqueID]
				light.updateState(event.State)
				go d.OnLightStateChange(d, event.UniqueID, event.State)
			}
		}
	}()

	return nil
}

func (d *Deconz) GetSensor(uniqueid string) (*Sensor, error) {
	if d.sensors[uniqueid] == nil {
		return nil, fmt.Errorf("sensor not found")
	}
	return d.sensors[uniqueid], nil
}
func (d *Deconz) GetLight(uniqueid string) (*Light, error) {
	if d.lights[uniqueid] == nil {
		return nil, fmt.Errorf("light not found")
	}
	return d.lights[uniqueid], nil
}
func (d *Deconz) updateSensors() error {
	res, err := d.client.Get(fmt.Sprintf("http://%s/api/%s/sensors", d.HubAddress, d.Username))
	if err != nil {
		return fmt.Errorf("failed to update sensors: %w", err)
	}

	defer res.Body.Close()

	var m map[string]*Sensor
	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for k, v := range m {
		v.Id = k
		v.Hub = d
		sensor := m[k]
		if d.sensors[sensor.UniqueID] == nil {
			d.sensors[sensor.UniqueID] = sensor
		} else {
			d.sensors[sensor.UniqueID].State = sensor.State
		}

		if d.OnSensorFound != nil {
			go d.OnSensorFound(d, sensor.UniqueID, sensor)
		}
	}

	return nil
}
func (d *Deconz) updateLights() error {
	res, err := d.client.Get(fmt.Sprintf("http://%s/api/%s/lights", d.HubAddress, d.Username))
	if err != nil {
		return fmt.Errorf("failed to update sensors: %w", err)
	}

	defer res.Body.Close()

	var m map[string]*Light
	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for k, v := range m {
		v.Id = k
		v.Hub = d
		d.lights[v.UniqueID] = v
		if d.OnLightFound != nil {
			go d.OnLightFound(d, v.UniqueID, v)
		}
	}

	return nil
}