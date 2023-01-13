package shelly

import (
	"fmt"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type ShellyDW2 struct {
	DeviceId   string
	mqttClient MQTT.Client
	mqttOpts   *MQTT.ClientOptions
}

func NewShellyDW2(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyDW2 {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyDW2{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}
	log.Printf("New ShellyDW2: %+v\n", s)
	return s
}

func (s ShellyDW2) Connect() {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (s ShellyDW2) Close() {
	s.mqttClient.Disconnect(disconnectQiesceTimeMs)
	log.Printf("%s: disconnected\n", s.DeviceName())
}

func (s ShellyDW2) DeviceName() string {
	return fmt.Sprintf("shellydw2-%s", s.DeviceId)
}

func (s ShellyDW2) baseTopic() string {
	return fmt.Sprintf("shellies/%s", s.DeviceName())
}

func (s ShellyDW2) SubscribeOpenState(openHandler func(), closeHandler func()) {
	topic := s.baseTopic() + "/sensor/state"

	openStateCallback := func(client MQTT.Client, message MQTT.Message) {
		if string(message.Payload()) == "open" {
			log.Printf("%s: window opened\n", s.DeviceName())
			openHandler()
		} else if string(message.Payload()) == "close" {
			log.Printf("%s: window closed\n", s.DeviceName())
			closeHandler()
		}
	}

	if token := s.mqttClient.Subscribe(topic, byte(qos), openStateCallback); token.Wait() &&
		token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("%s: subscribed to %s\n", s.DeviceName(), topic)
}
