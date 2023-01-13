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

type ShellyTRVInfo struct {
	Calibrated  bool                  `json:"calibrated"`
	Charger     bool                  `json:"charger"`
	PsMode      int                   `json:"ps_mode"`
	DbgFlags    int                   `json:"dbg_flags"`
	Thermostats []ShellyTRVThermostat `json:"thermostats"`
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
	log.Printf("%s: disconnected\n", s.deviceName())
}

func (s ShellyTRV) deviceName() string {
	return fmt.Sprintf("shellytrv-%s", s.DeviceId)
}

func (s ShellyTRV) baseTopic() string {
	return fmt.Sprintf("shellies/%s", s.deviceName())
}

func (s ShellyTRV) baseCommandTopic() string {
	return s.baseTopic() + "/thermostat/0/command"
}

func (s ShellyTRV) SetValve(valvePos float32) {
	log.Printf("%s: setting valve_pos to %f\n", s.deviceName(), valvePos)
	topic := s.baseCommandTopic() + "/valve_pos"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(valvePos))
	token.Wait()
}

func (s ShellyTRV) SetScheduleEnable(enable bool) {
	log.Printf("%s: setting schedule enable to %d\n", s.deviceName(), Btoi(enable))
	topic := s.baseCommandTopic() + "/schedule"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(Btoi(enable)))
	token.Wait()
}

func (s ShellyTRV) SetTargetTemperature(temperatureDegreeC float32) {
	log.Printf("%s: setting target temperature to %f °C\n", s.deviceName(), temperatureDegreeC)
	topic := s.baseCommandTopic() + "/target_t"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(temperatureDegreeC))
	token.Wait()
}

func (s ShellyTRV) SetExternalTemperature(temperatureDegreeC float32) {
	log.Printf("%s: setting external temperature to %f °C\n", s.deviceName(), temperatureDegreeC)
	topic := s.baseCommandTopic() + "/ext_t"
	token := s.mqttClient.Publish(topic, byte(qos), false, fmt.Sprint(temperatureDegreeC))
	token.Wait()
}

func (s ShellyTRV) pokeSettings() {
	log.Printf("%s: poking for settings\n", s.deviceName())

	topic := s.baseCommandTopic() + "/settings"
	token := s.mqttClient.Publish(topic, byte(qos), false, "")
	token.Wait()
}

func (s ShellyTRV) SubscribeStatus() {
	topic := s.baseTopic() + "/status"

	if token := s.mqttClient.Subscribe(topic, byte(qos), s.statusCallback); token.Wait() &&
		token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("%s: subscribed to status\n", s.deviceName())
}

func (s ShellyTRV) statusCallback(client MQTT.Client, message MQTT.Message) {
	log.Printf("%s: received message: %+v\n", s.deviceName(), message)
}

type ShellyTRVInfoCallback = func(info ShellyTRVInfo)

func (s ShellyTRV) SubscribeInfo(infoCallback ShellyTRVInfoCallback) {
	topic := s.baseTopic() + "/info"

	callback := func(client MQTT.Client, message MQTT.Message) {
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

	log.Printf("%s: subscribed to %s\n", s.deviceName(), topic)
}
