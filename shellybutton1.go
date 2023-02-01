package shelly

import (
	"fmt"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ShellyButton1 struct {
	ShellyDevice
}

type ShellyButton1InputEvent struct {
	Event    string `json:"event"`
	EventCnt int32  `json:"event_cnt"`
}

func NewShellyButton1(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyButton1 {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyButton1{ShellyDevice{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}}
	log.Debug().Str("DeviceName", s.DeviceName()).Msg("New ShellyButton1")
	return s
}

func (s ShellyButton1) Connect() {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error().
			Str("DeviceName", s.DeviceName()).
			Err(token.Error()).
			Msg("Error connecting to MQTT!")
	}
	log.Info().Str("DeviceName", s.DeviceName()).Msg("connected")
}

func (s ShellyButton1) Close() {
	s.mqttClient.Disconnect(disconnectQiesceTimeMs)
	log.Info().Str("DeviceName", s.DeviceName()).Msg("disconnected")
}

func (s ShellyButton1) DeviceName() string {
	return fmt.Sprintf("shellybutton1-%s", s.DeviceId)
}

func (s ShellyButton1) baseTopic() string {
	return fmt.Sprintf("shellies/%s", s.DeviceName())
}

func (s ShellyButton1) SubscribeBattery(batteryHandler func(float32)) {
	topic := s.baseTopic() + "/sensor/battery"
	batteryCallback := func(batteryStr string) {
		battery, err := strconv.ParseFloat(batteryStr, 32)
		if err != nil {
			log.Error().Str("powerStr", batteryStr).Msg("error parsing batteryStr as float32")
		}
		batteryHandler(float32(battery))
	}

	SubscribeStringHelper(s.mqttClient, topic, batteryCallback)
}

type ShellyButton1InputEventRawCallback = func(inputEvent ShellyButton1InputEvent)

func (s ShellyButton1) SubscribeInputEventRaw(
	inputEventCallback ShellyButton1InputEventRawCallback,
) {
	topic := s.baseTopic() + "/input_event/0"
	SubscribeJSONHelper(s.mqttClient, topic, inputEventCallback)
}

func (s ShellyButton1) SubscribeInputEvent(
	shortPressHandler func(),
	longPressHandler func(),
	doubleShortPressHandler func(),
	tripleShortPressHandler func(),
) {
	inputEventCallback := func(inputEvent ShellyButton1InputEvent) {
		switch inputEvent.Event {
		case "S":
			if shortPressHandler != nil {
				shortPressHandler()
			}
		case "L":
			if longPressHandler != nil {
				longPressHandler()
			}
		case "SS":
			if doubleShortPressHandler != nil {
				doubleShortPressHandler()
			}
		case "SSS":
			if tripleShortPressHandler != nil {
				tripleShortPressHandler()
			}
		default:
			log.Error().Str("inputEvent.Event", inputEvent.Event).Msg("unknown input event")
		}
	}

	s.SubscribeInputEventRaw(inputEventCallback)
}
