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

	plugS := shelly.NewShellyPlugS("EF6948", mqttOpts)
	plugS.Connect()
	defer plugS.Close()

	plugS.SubscribeRelayState(func() {
		log.Info().Msg("Switched on!")
	}, func() {
		log.Info().Msg("Switched off!")
	})

	var i = 0
	for {
		if i%2 == 0 {
			plugS.SwitchOn()
		} else {
			plugS.SwitchOff()
		}
		i++

		time.Sleep(time.Second * 5)
	}
}
