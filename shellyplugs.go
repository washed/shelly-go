package shelly

import (
	"fmt"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ShellyPlugS struct {
	ShellyDevice
}

func NewShellyPlugS(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyPlugS {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyPlugS{ShellyDevice{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}}
	log.Debug().Str("DeviceName", s.DeviceName()).Msg("New ShellyPlugS")
	return s
}

func (s ShellyPlugS) Connect() {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error().
			Str("DeviceName", s.DeviceName()).
			Err(token.Error()).
			Msg("Error connecting to MQTT!")
	}
	log.Info().Str("DeviceName", s.DeviceName()).Msg("connected")
}

func (s ShellyPlugS) Close() {
	s.mqttClient.Disconnect(disconnectQiesceTimeMs)
	log.Info().Str("DeviceName", s.DeviceName()).Msg("disconnected")
}

func (s ShellyPlugS) DeviceName() string {
	return fmt.Sprintf("shellyplug-s-%s", s.DeviceId)
}

func (s ShellyPlugS) baseTopic() string {
	return fmt.Sprintf("shellies/%s", s.DeviceName())
}

func (s ShellyPlugS) baseCommandTopic() string {
	return s.baseTopic() + "/relay/0/command"
}

func (s ShellyPlugS) SubscribeRelayState(onHandler func(), offHandler func()) {
	topic := s.baseTopic() + "/relay/0"
	relayStateCallback := func(relayState string) {
		if relayState == "on" {
			onHandler()
		} else if relayState == "off" {
			offHandler()
		} else {
			log.Error().Str("relayState", relayState).Msg("received unknown relayState value")
		}
	}

	SubscribeStringHelper(s.mqttClient, topic, relayStateCallback)
}

func (s ShellyPlugS) SubscribePower(powerHandler func(float32)) {
	topic := s.baseTopic() + "/relay/0/power"
	powerCallback := func(powerStr string) {
		power, err := strconv.ParseFloat(powerStr, 32)
		if err != nil {
			log.Error().Str("powerStr", powerStr).Msg("error parsing powerStr as float32")
		}
		powerHandler(float32(power))
	}

	SubscribeStringHelper(s.mqttClient, topic, powerCallback)
}

func (s ShellyPlugS) switchRelay(relayState bool) {
	log.Info().
		Str("DeviceName", s.DeviceName()).
		Msg("switching on")
	topic := s.baseCommandTopic()

	command := "off"
	if relayState {
		command = "on"
	}

	token := s.mqttClient.Publish(topic, byte(qos), false, command)
	token.Wait()
}

func (s ShellyPlugS) SwitchOn() {
	s.switchRelay(true)

}

func (s ShellyPlugS) SwitchOff() {
	s.switchRelay(false)
}
