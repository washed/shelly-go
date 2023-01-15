package main

import (
	"encoding/json"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	broker   = os.Getenv("MQTT_BROKER_URL")
	user     = os.Getenv("MQTT_BROKER_USERNAME")
	password = os.Getenv("MQTT_BROKER_PASSWORD")
)

func pokeAnnounce(mqttClient MQTT.Client) {
	log.Info().
		Msg("Poking for shelly announce")
	topic := "shellies/command"
	token := mqttClient.Publish(topic, byte(0), false, "announce")
	token.Wait()
}

type ShellyAnnounce struct {
	ID    string `json:"id"`
	Model string `json:"model"`
	MAC   string `json:"mac"`
	IP    string `json:"ip"`
	NewFW bool   `json:"new_fw"`
	FWVer string `json:"fw_ver"`
}

func main() {
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.000Z07:00"
	log.Logger = log.Output(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano},
	)

	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(broker)
	mqttOpts.SetUsername(user)
	mqttOpts.SetPassword(password)

	mqttClient := MQTT.NewClient(mqttOpts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Error().
			Err(token.Error()).
			Msg("Error connecting to MQTT!")
		return
	}
	log.Info().Msg("connected")

	defer mqttClient.Disconnect(250)

	topic := "shellies/announce"

	callback := func(client MQTT.Client, message MQTT.Message) {

		announce := ShellyAnnounce{}
		err := json.Unmarshal(message.Payload(), &announce)
		if err != nil {
			log.Error().
				Str("message.Payload", string(message.Payload())).
				Err(err).
				Msg("Error unmarshalling ShellyAnnounce")
			return
		}

		log.Info().
			Interface("announce", announce).
			Msg("Received ShellyAnnounce")
	}

	if token := mqttClient.Subscribe(topic, byte(0), callback); token.Wait() &&
		token.Error() != nil {
		log.Error().
			Str("topic", topic).
			Err(token.Error()).
			Msg("Error subscribing!")
		return
	}

	log.Info().
		Str("topic", topic).
		Msg("Subscribed!")

	pokeAnnounce(mqttClient)

	for {
		time.Sleep(time.Second * 10)
	}
}
