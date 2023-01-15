package shelly

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ShellyDW2 struct {
	ShellyDevice
}

type ShellyDW2Sensor struct {
	State   string `json:"state"`
	IsValid bool   `json:"is_valid"`
}

type ShellyDW2Accel struct {
	Tilt      int `json:"tilt"`
	Vibration int `json:"vibration"`
}

type ShellyDW2Info struct {
	Sensor ShellyDW2Sensor `json:"sensor"`
	Bat    ShellyInfoBat   `json:"bat"`
	Tmp    ShellyInfoTmp   `json:"tmp"`
	Lux    ShellyInfoLux   `json:"lux"`
	Accel  ShellyDW2Accel  `json:"accel"`
}

/*
Implement the rest the remaining info fields if necessary?
{
    "wifi_sta": {
        "connected": true,
        "ssid": "",
        "ip": "192.168.178.86",
        "rssi": -37
    },
    "cloud": {
        "enabled": true,
        "connected": false
    },
    "mqtt": {
        "connected": true
    },
    "time": "",
    "unixtime": 0,
    "serial": 1,
    "has_update": false,
    "mac": "485519C92B94",
    "cfg_changed_cnt": 0,
    "actions_stats": {
        "skipped": 0
    },
    "is_valid": true,
    "tmp": {
        "tC": 17.30,
        "tF": 63.14,
    },
    "accel": {
        "tilt": 8,
        "vibration": -1
    },
    "act_reasons": [
        "sensor"
    ],
    "sensor_error": 0,
    "update": {
        "status": "unknown",
        "has_update": false,
        "new_version": "",
        "old_version": "20220209-093605/v1.11.8-g8c7bb8d"
    },
    "ram_total": 51352,
    "ram_free": 40544,
    "fs_size": 233681,
    "fs_free": 154867,
    "uptime": 1
}
*/

func NewShellyDW2(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyDW2 {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyDW2{ShellyDevice{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}}
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
	openStateCallback := func(windowState string) {
		if windowState == "open" {
			openHandler()
		} else if windowState == "close" {
			closeHandler()
		} else {
			log.Error().Str("state", windowState).Msg("received unknown state value")
		}
	}

	SubscribeStringHelper(s.mqttClient, topic, openStateCallback)
}

type ShellyDW2InfoCallback = func(info ShellyDW2Info)

func (s ShellyDW2) SubscribeInfo(infoCallback ShellyDW2InfoCallback) {
	topic := s.baseTopic() + "/info"
	SubscribeJSONHelper(s.mqttClient, topic, infoCallback)
}
