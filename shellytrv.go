package shelly

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

type ShellyTRV struct {
	DeviceId   string
	mqttClient MQTT.Client
	mqttOpts   *MQTT.ClientOptions
}

type ShellyTRVThermostat struct {
	Pos             float32          `json:"pos"`
	Schedule        bool             `json:"schedule"`
	ScheduleProfile int              `json:"schedule_profile"`
	BoostMinutes    int              `json:"boost_minutes"`
	TargetT         ShellyTRVTargetT `json:"target_t"`
	Tmp             ShellyTRVTmp     `json:"tmp"`
}

type ShellyTRVTargetT struct {
	Enabled bool    `json:"enabled"`
	Value   float32 `json:"value"`
	Units   string  `json:"units"`
}

type ShellyTRVTmp struct {
	Value   float32 `json:"value"`
	Units   string  `json:"units"`
	IsValid bool    `json:"is_valid"`
}

type ShellyTRVBat struct {
	Value   int     `json:"value"`
	Voltage float32 `json:"voltage"`
}

type ShellyTRVInfo struct {
	Calibrated  bool                  `json:"calibrated"`
	Charger     bool                  `json:"charger"`
	PsMode      int                   `json:"ps_mode"`
	DbgFlags    int                   `json:"dbg_flags"`
	Thermostats []ShellyTRVThermostat `json:"thermostats"`
	Bat         ShellyTRVBat          `json:"bat"`
}

/*
Implement the rest the remaining info fields if necessary?
{
    "wifi_sta": {
        "connected": true,
        "ssid": "wpd.wlan-2.4GHz",
        "ip": "192.168.178.123",
        "rssi": -33
    },
    "cloud": {
        "enabled": false,
        "connected": false
    },
    "mqtt": {
        "connected": true
    },
    "time": "17:42",
    "unixtime": 1673628121,
    "serial": 0,
    "has_update": false,
    "mac": "60A423DAE8DE",
    "cfg_changed_cnt": 0,
    "actions_stats": {
        "skipped": 0
    },
    "bat": {
        "value": 99,
        "voltage": 3.989
    },
    "update": {
        "status": "unknown",
        "has_update": false,
        "new_version": "20220811-152343/v2.1.8@5afc928c",
        "old_version": "20220811-152343/v2.1.8@5afc928c",
        "beta_version": null
    },
    "ram_total": 97280,
    "ram_free": 22488,
    "fs_size": 65536,
    "fs_free": 59416,
    "uptime": 318520,
    "fw_info": {
        "device": "shellytrv-60A423DAE8DE",
        "fw": "20220811-152343/v2.1.8@5afc928c"
    },
}
*/

type ShellyTRVStatus struct {
	TargetT           ShellyTRVTargetT `json:"target_t"`
	Tmp               ShellyTRVTmp     `json:"tmp"`
	TemperatureOffset float32          `json:"temperature_offset"`
	Bat               float32          `json:"bat"`
}

func NewShellyTRV(deviceId string, mqttOpts *MQTT.ClientOptions) ShellyTRV {
	client := MQTT.NewClient(mqttOpts)
	s := ShellyTRV{DeviceId: deviceId, mqttClient: client, mqttOpts: mqttOpts}
	log.Printf("New ShellyTRV: %+v\n", s)
	return s
}

func (s ShellyTRV) Connect() {
	if token := s.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func (s ShellyTRV) Close() {
	s.mqttClient.Disconnect(disconnectQiesceTimeMs)
	log.Printf("%s: disconnected\n", s.DeviceName())
}

func (s ShellyTRV) DeviceName() string {
	return fmt.Sprintf("shellytrv-%s", s.DeviceId)
}

func (s ShellyTRV) baseTopic() string {
	return fmt.Sprintf("shellies/%s", s.DeviceName())
}

func (s ShellyTRV) baseCommandTopic() string {
	return s.baseTopic() + "/thermostat/0/command"
}

func (s ShellyTRV) SetValve(valvePos float32) {
	log.Printf("%s: setting valve_pos to %f\n", s.DeviceName(), valvePos)
	topic := s.baseCommandTopic() + "/valve_pos"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(valvePos))
	token.Wait()
}

func (s ShellyTRV) SetScheduleEnable(enable bool) {
	log.Printf("%s: setting schedule enable to %d\n", s.DeviceName(), Btoi(enable))
	topic := s.baseCommandTopic() + "/schedule"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(Btoi(enable)))
	token.Wait()
}

func (s ShellyTRV) SetTargetTemperature(temperatureDegreeC float32) {
	log.Printf("%s: setting target temperature to %f °C\n", s.DeviceName(), temperatureDegreeC)
	topic := s.baseCommandTopic() + "/target_t"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(temperatureDegreeC))
	token.Wait()
}

func (s ShellyTRV) SetExternalTemperature(temperatureDegreeC float32) {
	log.Printf("%s: setting external temperature to %f °C\n", s.DeviceName(), temperatureDegreeC)
	topic := s.baseCommandTopic() + "/ext_t"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(temperatureDegreeC))
	token.Wait()
}

func (s ShellyTRV) pokeSettings() {
	log.Printf("%s: poking for settings\n", s.DeviceName())

	topic := s.baseCommandTopic() + "/settings"
	token := s.mqttClient.Publish(topic, byte(qos), false, "")
	token.Wait()
}

type ShellyTRVStatusCallback = func(status ShellyTRVStatus)

func (s ShellyTRV) SubscribeStatus(statusCallback ShellyTRVStatusCallback) {
	topic := s.baseTopic() + "/status"

	callback := func(client MQTT.Client, message MQTT.Message) {
		log.Printf("%s: received message: %+v\n", s.DeviceName(), string(message.Payload()))
		status := ShellyTRVStatus{}
		err := json.Unmarshal(message.Payload(), &status)
		if err != nil {
			log.Printf("Error unmarshaling message '%+v'! Error: '%+v'.\n", message, err)
		}
		statusCallback(status)
	}

	if token := s.mqttClient.Subscribe(topic, byte(qos), callback); token.Wait() &&
		token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("%s: subscribed to %s\n", s.DeviceName(), topic)
}

type ShellyTRVInfoCallback = func(info ShellyTRVInfo)

func (s ShellyTRV) SubscribeInfo(infoCallback ShellyTRVInfoCallback) {
	topic := s.baseTopic() + "/info"

	callback := func(client MQTT.Client, message MQTT.Message) {
		log.Printf("%s: received message: %+v\n", s.DeviceName(), string(message.Payload()))
		info := ShellyTRVInfo{}
		err := json.Unmarshal(message.Payload(), &info)
		if err != nil {
			log.Printf("Error unmarshaling message '%+v'! Error: '%+v'.\n", message, err)
		}
		infoCallback(info)
	}

	if token := s.mqttClient.Subscribe(topic, byte(qos), callback); token.Wait() &&
		token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("%s: subscribed to %s\n", s.DeviceName(), topic)
}

func (s ShellyTRV) SubscribeAll() {
	topic := s.baseTopic() + "/#"

	callback := func(client MQTT.Client, message MQTT.Message) {
		log.Printf(
			"%s: received topic: %+v message: %+v\n",
			s.DeviceName(),
			string(message.Topic()),
			string(message.Payload()),
		)
	}

	if token := s.mqttClient.Subscribe(topic, byte(qos), callback); token.Wait() &&
		token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("%s: subscribed to %s\n", s.DeviceName(), topic)
}
