package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Salus struct {
	credentials Credentials
	token       string
	devices     map[string]device
}

type Credentials struct {
	email    string
	password string
}

type device struct {
	id     string
	values deviceValues
}

type deviceValues struct {
	Temperature        float64 `json:"CH1currentRoomTemp,string"`
	SetPoint           float64 `json:"CH1currentSetPoint,string"`
	HeaterStatusString string  `json:"CH1heatOnOffStatus"`
	HeaterStatus       bool
}

func New(credentials Credentials) *Salus {
	s := Salus{
		credentials: credentials,
		devices:     make(map[string]device),
	}

	s.initTokenAndDeviceIds()

	for i, d := range s.devices {
		d.values = s.initDevice(d)
		s.devices[i] = d
	}

	return &s
}

func (s *Salus) GetTemperature(deviceCode string) float64 {
	return s.devices[deviceCode].values.Temperature
}

func (s *Salus) GetSetPoint(deviceCode string) float64 {
	return s.devices[deviceCode].values.SetPoint
}

func (s *Salus) GetIsHeating(deviceCode string) bool {
	return s.devices[deviceCode].values.HeaterStatus
}

func (s *Salus) initDevice(d device) deviceValues {
	resp, err := http.Get(fmt.Sprintf("https://salus-it500.com/public/ajax_device_values.php?devId=%s&token=%s&_=%d", d.id, s.token, time.Now().Unix()))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	dv := deviceValues{}
	json.Unmarshal(bodyBytes, &dv)

	dv.HeaterStatus = false
	if dv.HeaterStatusString == "1" {
		dv.HeaterStatus = true
	}

	return dv
}

func (s *Salus) initTokenAndDeviceIds() {
	client := &http.Client{}

	resp, err := client.Get("https://salus-it500.com/public/login.php?")
	if err != nil {
		panic(err)
	}
	cookie := strings.Split(resp.Header["Set-Cookie"][0], ";")[0]

	form := url.Values{}
	form.Add("IDemail", s.credentials.email)
	form.Add("password", s.credentials.password)
	form.Add("login", "Login")

	req, err := http.NewRequest("POST", "https://salus-it500.com/public/login.php?", strings.NewReader(form.Encode()))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	devIdExists, err := regexp.Match("<input name=\"devId\"", bodyBytes)
	if err != nil {
		panic(err)
	}
	if !devIdExists {
		// failed logins return 200 - checking for container with deviceId to indicate success
		panic("Login failed")
	}

	req, err = http.NewRequest("GET", "https://salus-it500.com/public/devices.php", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Cookie", cookie)

	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("control\\.php\\?devId=(\\d+)\">([A-Z0-9]*)")

	for _, sb := range re.FindAllSubmatch(bodyBytes, -1) {
		s.devices[string(sb[2])] = device{
			id: string(sb[1]),
		}
	}

	re = regexp.MustCompile("<input id=\"token\" name=\"token\" type=\"hidden\" value=\"(\\d+-[a-zA-Z0-9]+)\" />")
	s.token = string(re.FindSubmatch(bodyBytes)[1])
}
