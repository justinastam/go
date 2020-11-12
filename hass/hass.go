package hass

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Hass struct {
	baseUrl string
	bearer string
}

type SensorStateAttributes struct {
	FriendlyName		string	`json:"friendly_name"`
	Icon				string 	`json:"icon"`
	UnitOfMeasurement	string	`json:"unit_of_measurement"`
}

type SensorStateContext struct {
	Id			string	`json:"id"`
	ParentId	string	`json:"parent_id"`
}

type SensorState struct {
	Attributes		SensorStateAttributes	`json:"attributes"`
	Context			SensorStateContext	`json:"context"`
	EntityId		string	`json:"entity_id"`
	LastChanged   	string  `json:"last_changed"`
	LastUpdated   	string	`json:"last_updated"`
	State			string  `json:"state"`
}

func New(baseUrl string, bearer string) Hass {
	hass := Hass{
		baseUrl: baseUrl,
		bearer: bearer,
	}

	return hass
}

func (h Hass) GetSensorState(sensorId string) SensorState {
	var url = h.baseUrl + "api/states/" + sensorId

	resp := h.doRequest(url)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var sensorState SensorState
	json.Unmarshal(bodyBytes, &sensorState)

	return sensorState
}

func (h Hass) GetSensorHistory(sensorId string, ts time.Time) [][]SensorState {
	var url = h.baseUrl + "api/history/period/" + ts.Format("2006-01-02T15:04:05-07:00") + "?filter_entity_id=" + sensorId

	resp := h.doRequest(url)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var sensorStates [][]SensorState
	json.Unmarshal(bodyBytes, &sensorStates)

	if len(sensorStates) == 0 {
		return sensorStates
	}

	// fix remove wrong values
	for k, state := range sensorStates[0] {
		if !isNumeric(state.State) {
			removeStateFromSlice(sensorStates[0], k)
		}
	}

	return sensorStates
}

func (h Hass) doRequest(url string) *http.Response {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Authorization", "Bearer " + h.bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	return resp
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func removeStateFromSlice(slice []SensorState, s int) []SensorState {
	return append(slice[:s], slice[s+1:]...)
}
