package main

import (
	"log"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/washed/shelly-go"
)

var (
	broker   = os.Getenv("MQTT_BROKER_URL")
	user     = os.Getenv("MQTT_BROKER_USERNAME")
	password = os.Getenv("MQTT_BROKER_PASSWORD")
)

func infoCallback(info shelly.ShellyTRVInfo) {
	log.Printf("Received info: %+v\n", info)
}

func statusCallback(status shelly.ShellyTRVStatus) {
	log.Printf("Received status: %+v\n", status)
}

func main() {
	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(broker)
	mqttOpts.SetUsername(user)
	mqttOpts.SetPassword(password)

	trv := shelly.NewShellyTRV("60A423DAE8DE", mqttOpts)
	trv.Connect()
	defer trv.Close()

	trv.SubscribeInfo(infoCallback)
	trv.SubscribeStatus(statusCallback)

	for {
		time.Sleep(time.Second * 10)
	}
}
