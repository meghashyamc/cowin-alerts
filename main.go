package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

type VaccineAvailability struct {
	Centers []Center `json:"centers"`
}

type Center struct {
	Name      string    `json:"name"`
	BlockName string    `json:"block_name"`
	Pincode   int       `json:"pincode"`
	Sessions  []Session `json:"sessions"`
}

type Session struct {
	Date                    string   `json:"date"`
	MinAgeLimit             int      `json:"min_age_limit"`
	Slots                   []string `json:"slots"`
	Vaccine                 string   `json:"vaccine"`
	AvailableCapacity       int      `json:"available_capacity"`
	AvailalableCapaityDose1 int      `json:"available_capacity_dose1"`
	AvailalableCapaityDose2 int      `json:"available_capacity_dose2"`
}

var triggered bool

func main() {

	if err := godotenv.Load(); err != nil {
		log.Info("Could not load env variables: ", err)
	}

	centerCodes := strings.Split(os.Getenv("COWIN_CENTER_CODES"), ",")

outerloop:
	for {
		if triggered {
			triggered = false
			//If an alert is triggered, trigger the next one after 30 minutes
			time.Sleep(30 * time.Minute)
		} else {
			time.Sleep(1 * time.Minute)
		}

		for _, centerCode := range centerCodes {

			availableCentres, err := callCowinSite(centerCode)
			if err != nil {
				continue outerloop
			}
			if len(availableCentres) > 0 {
				log.Info("Centres available, trying to trigger alert")
				triggerAlert()
				triggered = true
			}
		}

		log.Info("No centres available right now")

	}

}

func callCowinSite(districtCode string) ([]Center, error) {

	url := os.Getenv("COWIN_DISTRICT_CODE_URL") + districtCode + "&date=" + time.Now().Format("02-01-2006")
	client := http.Client{Timeout: time.Second * 30}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error("Could not call Cowin API:", err.Error())

		return nil, err
	}
	req.Header.Set("User-Agent", os.Getenv("COWIN_REQUEST_USER_AGENT"))
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Did not receive successful response from Cowin API:", err.Error())
		return nil, errors.New("Did not receive successful response from Cowin API")
	}
	if resp.StatusCode != 200 {
		log.Error("Response code from Cowin API was not 200, it was", resp.StatusCode)
	}

	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Could not read response received from Cowin API:", err.Error())
		return nil, errors.New("Could not read response body")
	}
	if resp.StatusCode != 200 {
		log.Error("Response:", string(respBytes))
		return nil, errors.New("Received a response whose status code was not 200")
	}

	vaccineAvailability := VaccineAvailability{}

	if err := json.Unmarshal(respBytes, &vaccineAvailability); err != nil {
		log.Error("Could not unmarshal response received from Cowin API:", err.Error())

		return nil, errors.New("Could not unmarshal gotten response")

	}

	availableCenters := []Center{}

	for _, center := range vaccineAvailability.Centers {

		for _, session := range center.Sessions {

			if session.MinAgeLimit == 18 && session.AvailableCapacity > 0 && session.AvailalableCapaityDose1 > 0 {
				availableCenters = append(availableCenters, center)
				break
			}

		}
	}

	return availableCenters, nil

}

func triggerAlert() {

	postBody := `{"routing_key":"` + os.Getenv("PAGER_DUTY_ROUTING_KEY") + `","event_action":"trigger","payload":{"summary":"Center available! Check the Cowin website or run the program immediately!","severity":"critical","source":"my app"}}`
	requestBody := bytes.NewBuffer([]byte(postBody))
	resp, err := http.Post(os.Getenv("PAGER_DUTY_ALERT_URL"), "application/json", requestBody)
	if err != nil {
		log.Error("Pager Duty related error:", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Pager Duty related error:", err)
	}
	log.Info("Pager Duty response:", string(respBytes))
}
