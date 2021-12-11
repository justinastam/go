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
	credentials  Credentials
	token        string
	deviceId     string
	deviceValues DeviceValues
}

type Credentials struct {
	email    string
	password string
}

type DeviceValues struct {
	Temperature        float64 `json:"CH1currentRoomTemp,string"`
	SetPoint           float64 `json:"CH1currentSetPoint,string"`
	HeaterStatusString string  `json:"CH1heatOnOffStatus"`
	HeaterStatus       bool
	Initiated          bool
}

//// @todo delete
//func main() {
//	salus := New(getCredentials())
//	salus.GetTemperature()
//}
//
//// @todo delete
//func getCredentials() Credentials {
//	return Credentials{email: "email", password: "pass"}
//}

func New(credentials Credentials) Salus {
	s := Salus{
		credentials: credentials,
	}

	s.InitDeviceValues()

	return s
}

func (s *Salus) GetTemperature() float64 {
	return s.deviceValues.Temperature
}

func (s *Salus) GetSetPoint() float64 {
	return s.deviceValues.SetPoint
}

func (s *Salus) GetIsHeating() bool {
	return s.deviceValues.HeaterStatus
}

func (s *Salus) InitDeviceValues() {
	if s.deviceValues.Initiated {
		return
	}

	s.InitTokenAndDeviceId()

	url := fmt.Sprintf("https://salus-it500.com/public/ajax_device_values.php?devId=%s&token=%s&_=%d", s.deviceId, s.token, time.Now().Unix())
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	dv := DeviceValues{}
	json.Unmarshal(bodyBytes, &dv)
	dv.HeaterStatus = false
	if dv.HeaterStatusString == "1" {
		dv.HeaterStatus = true
	}
	dv.Initiated = true

	s.deviceValues = dv
}

func (s *Salus) InitTokenAndDeviceId() {
	if s.token != "" && s.deviceId != "" {
		return
	}

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

	re := regexp.MustCompile("control\\.php\\?devId=(\\d+)")
	s.deviceId = string(re.FindSubmatch(bodyBytes)[1])

	re = regexp.MustCompile("<input id=\"token\" name=\"token\" type=\"hidden\" value=\"(\\d+-[a-zA-Z0-9]+)\" />")
	s.token = string(re.FindSubmatch(bodyBytes)[1])
}
