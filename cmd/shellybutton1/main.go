package main

import (
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/washed/shelly-go"
)

var (
	broker   = os.Getenv("MQTT_BROKER_URL")
	user     = os.Getenv("MQTT_BROKER_USERNAME")
	password = os.Getenv("MQTT_BROKER_PASSWORD")
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log.Logger = log.Output(
		zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339Nano},
	)

	mqttOpts := MQTT.NewClientOptions()
	mqttOpts.AddBroker(broker)
	mqttOpts.SetUsername(user)
	mqttOpts.SetPassword(password)

	button1 := shelly.NewShellyButton1("3C6105E51C74", mqttOpts)
	button1.Connect()
	defer button1.Close()

	button1.SubscribeInputEventRaw(func(inputEvent shelly.ShellyButton1InputEvent) {
		log.Info().Interface("inputEvent", inputEvent).Msg("received input event")
	})

	button1.SubscribeInputEvent(func() {
		log.Info().Str("type", "short press").Msg("received input event")
	}, func() {
		log.Info().Str("type", "long press").Msg("received input event")
	}, func() {
		log.Info().Str("type", "double short press").Msg("received input event")
	}, func() {
		log.Info().Str("type", "triple short press").Msg("received input event")
	})

	button1.SubscribeBattery(func(battery float32) {
		log.Info().Float32("battery", battery).Msg("received battery status")
	})

	for {
		time.Sleep(time.Second * 5)
	}
}
