package shelly

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type ShellyDevice struct {
	DeviceId   string
	mqttClient MQTT.Client
	mqttOpts   *MQTT.ClientOptions
}
