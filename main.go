package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi"
	"github.com/taiyoh/go-embd-bme280"
)

type CPUData struct {
	Temperature float64 `json:"temperature"`
}

type BME280Data struct {
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Humidity    float64 `json:"humidity"`
}

type Result struct {
	CPU    CPUData    `json:"cpu"`
	BME280 BME280Data `json:"bme280"`
}

func readCPUData() (*CPUData, error) {
	if cpuTempRaw, err := ioutil.ReadFile("/sys/class/thermal/thermal_zone0/temp"); err != nil {
		return nil, fmt.Errorf("Failed to read CPU temperature: %s", err)
	} else {
		cpuTemp, err := strconv.ParseFloat(strings.TrimSpace(string(cpuTempRaw)), 64)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse CPU temperature: %s", err)
		} else {
			return &CPUData{cpuTemp / 1000.0}, nil
		}
	}
}

func readBME280Data() (*BME280Data, error) {
	if err := embd.InitI2C(); err != nil {
		return nil, fmt.Errorf("Failed to read BME280 data: %s", err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	opt := bme280.NewOpt()
	bme280, err := bme280.New(bus, opt)
	if err != nil {
		return nil, fmt.Errorf("Failed to read BME280 data: %s", err)
	}

	data, err := bme280.Read()
	if err != nil {
		return nil, fmt.Errorf("Failed to read BME280 data: %s", err)
	} else {
		return &BME280Data{data[0], data[1] / 100, data[2]}, nil
	}
}

func convertJSON(result *Result) (string, error) {
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("Failed to convert to JSON: %s", err)
	}
	return string(jsonBytes), nil
}

func main() {
	result := Result{}

	if cpuData, err := readCPUData(); err != nil {
		panic(err)
	} else {
		result.CPU = *cpuData
	}

	if bme280Data, err := readBME280Data(); err != nil {
		panic(err)
	} else {
		result.BME280 = *bme280Data
	}

	if json, err := convertJSON(&result); err != nil {
		panic(err)
	} else {
		fmt.Println(json)
	}
}
