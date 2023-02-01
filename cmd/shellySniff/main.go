package main

import (
	"os"
	"strings"
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

	topic := "shellies/#"
	filter := "shellybutton"

	callback := func(client MQTT.Client, message MQTT.Message) {
		if strings.Contains(message.Topic(), filter) {
			log.Info().
				Str("topic", message.Topic()).
				Str("payload", string(message.Payload())).
				Msg("Received shelly message")
		}
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
		Str("topic_filter", filter).
		Msg("subscribed")

	for {
		time.Sleep(time.Second * 10)
	}
}
