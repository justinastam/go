package salus

import (
	"os"
	"strconv"
	"testing"
)

func TestGetTemperature(t *testing.T) {
	salus := New(getCredentials())

	expTemp, err := strconv.ParseFloat(os.Getenv("EXP_TEMP"), 64)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	salusTemp := salus.GetTemperature("STA10072968")
	if salusTemp != expTemp {
		t.Errorf("Temperature incorrect, got: %f, want: %f.", salusTemp, expTemp)
	}
}

func TestGetSetPoint(t *testing.T) {
	salus := New(getCredentials())

	expSP, err := strconv.ParseFloat(os.Getenv("EXP_SET_POINT"), 64)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}

	setPoint := salus.GetSetPoint("STA00074484")
	if setPoint != expSP {
		t.Errorf("Set point incorrect, got: %f, want: %f.", setPoint, expSP)
	}
}

func TestIsHeating(t *testing.T) {
	salus := New(getCredentials())

	expIsHeating := false
	if os.Getenv("EXP_IS_HEATING") == "1" {
		expIsHeating = true
	} else if os.Getenv("EXP_IS_HEATING") != "0" {
		panic("Unknown expectation for heating")
	}

	heating := salus.GetIsHeating("STA10072968")
	if heating != expIsHeating {
		t.Errorf("Heater status incorrect")
	}
}

func getCredentials() Credentials {
	return Credentials{email: os.Getenv("SALUS_EMAIL"), password: os.Getenv("SALUS_PASSWORD")}
}
