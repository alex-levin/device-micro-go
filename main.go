package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Device struct {
	// https://www.digitalocean.com/community/tutorials/how-to-use-struct-tags-in-go
	ID      int    `json:"id"`
	ESN     string `json:"esn"`
	Status  string `json:"status"`
	Address string `json:"address"`
	Name    string `json:"name"`
}

var deviceList []Device

// https://www.digitalocean.com/community/tutorials/understanding-init-in-go
// In Go, the predefined init() function sets off a piece of code to run before any other part of your package
func init() {
	devicesJSON := `[
		{
			"id": 1,
			"esn": "A0000001",
			"status": "Online",
			"address": "127.0.1.1",
			"name": "Charles Darnay"
		},
		{
			"id": 2,
			"esn": "A0000003",
			"status": "Online",
			"address": "127.0.1.2",
			"name": "Sydney Carton"
		},
		{
			"id": 3,
			"esn": "A0000003",
			"status": "Offline",
			"address": "127.0.1.3",
			"name": "Lucie Manette"
		},
		{
			"id": 4,
			"esn": "A0000004",
			"status": "Online",
			"address": "127.0.1.4",
			"name": "Miss Pross"
		},
		{
			"id": 5,
			"esn": "A0000005",
			"status": "Online",
			"address": "127.0.1.5",
			"name": "Jarvis Lorry"
		}	
	]`
	err := json.Unmarshal([]byte(devicesJSON), &deviceList)
	if err != nil {
		log.Fatal(err)
	}
}

func getNextID() int {
	highestID := -1
	// _ is the blank identifier (unused variable) - the deviceList slice index
	// https://golang.org/doc/effective_go
	for _, device := range deviceList {
		if highestID < device.ID {
			highestID = device.ID
		}
	}
	return highestID + 1
}

func findDeviceByID(deviceID int) (*Device, int) {
	for i, device := range deviceList {
		if device.ID == deviceID {
			return &device, i
		}
	}
	return nil, 0
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request URI: ", r.RequestURI)
	switch r.Method {
		case http.MethodGet:
			devicesJSON, err := json.Marshal(deviceList)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(devicesJSON)
		case http.MethodPost:
			// add a new device to the list
			var newDevice Device
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err = json.Unmarshal(bodyBytes, &newDevice)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			// We want detabase to assign the ID
			if newDevice.ID != 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			newDevice.ID = getNextID()
			deviceList = append(deviceList, newDevice)
			// 201
			w.WriteHeader(http.StatusCreated)
			return
	}
}

func deviceHandler(w http.ResponseWriter, r *http.Request) {
	// URL: localhost:5000/devices/123
	urlPathSegments := strings.Split(r.URL.Path, "devices/")
	deviceID, err := strconv.Atoi(urlPathSegments[len(urlPathSegments)-1])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	device, listItemIndex := findDeviceByID(deviceID)
	if device == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch r.Method {
	case http.MethodGet:
		deviceJSON, err := json.Marshal(device)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(deviceJSON)
	case http.MethodPut:
		// update device
		var updateDevice Device
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(bodyBytes, &updateDevice)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if updateDevice.ID != deviceID {
			// 400
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		device = &updateDevice
		deviceList[listItemIndex] = *device
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func main() {
	// https://golang.org/pkg/net/http/#HandleFunc
	http.HandleFunc("/devices", devicesHandler)
	http.HandleFunc("/devices/", deviceHandler)
	// http://localhost:8080/devices
	http.ListenAndServe((":8080"), nil)
}
