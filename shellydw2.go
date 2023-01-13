package shelly

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ShellyDW2 struct {
	DeviceId   string
	mqttClient MQTT.Client
	mqttOpts   *MQTT.ClientOptions
}

func NewShellyDW2(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyDW2 {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyDW2{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}
	log.Debug().Str("DeviceName", s.DeviceName()).Msg("New ShellyDW2")
	return s
}

func (s ShellyDW2) Connect() {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error().
			Str("DeviceName", s.DeviceName()).
			Err(token.Error()).
			Msg("Error connecting to MQTT!")
	}
	log.Info().Str("DeviceName", s.DeviceName()).Msg("connected")
}

func (s ShellyDW2) Close() {
	s.mqttClient.Disconnect(disconnectQiesceTimeMs)
	log.Info().Str("DeviceName", s.DeviceName()).Msg("disconnected")
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
		windowState := string(message.Payload())
		log.Info().
			Str("DeviceName", s.DeviceName()).
			Str("windows state", windowState).
			Msg("window state changed")
		if windowState == "open" {
			openHandler()
		} else if windowState == "close" {
			closeHandler()
		}
	}

	if token := s.mqttClient.Subscribe(topic, byte(qos), openStateCallback); token.Wait() &&
		token.Error() != nil {
		log.Error().
			Str("DeviceName", s.DeviceName()).
			Str("topic", topic).
			Err(token.Error()).
			Msg("Error subscribing!")
		return
	}

	log.Info().
		Str("DeviceName", s.DeviceName()).
		Str("topic", topic).
		Msg("Subscribed!")
}
