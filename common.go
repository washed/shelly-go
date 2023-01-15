package shelly

import (
	"encoding/json"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type ShellyInfoBat struct {
	Value   int     `json:"value"`
	Voltage float32 `json:"voltage"`
}

type ShellyInfoTmp struct {
	Value   float32 `json:"value"`
	Units   string  `json:"units"`
	IsValid bool    `json:"is_valid"`
}

type ShellyInfoLux struct {
	Value        float32 `json:"value"`
	Illumination string  `json:"illumination"`
	IsValid      bool    `json:"is_valid"`
}

func logMessage(message MQTT.Message) {
	log.Debug().
		Str("message.Topic", string(message.Topic())).
		Str("message.Payload", string(message.Payload())).
		Msg("received message")
}

func checkedJSONUnmarshal[T ShellyDW2Info | ShellyTRVInfo | ShellyTRVStatus](
	message MQTT.Message,
	out *T,
) error {
	err := json.Unmarshal(message.Payload(), out)
	if err != nil {
		log.Error().
			Str("message.Topic", string(message.Topic())).
			Str("message.Payload", string(message.Payload())).
			Err(err).
			Msg("Error unmarshalling message!")
		return err
	}

	return nil
}

func checkedSubscribe(
	mqttClient MQTT.Client,
	topic string,
	callback func(client MQTT.Client, message MQTT.Message),
) error {
	if token := mqttClient.Subscribe(topic, byte(qos), callback); token.Wait() &&
		token.Error() != nil {
		log.Error().
			Str("topic", topic).
			Err(token.Error()).
			Msg("Error subscribing!")
		return token.Error()
	}

	log.Info().
		Str("topic", topic).
		Msg("Subscribed!")

	return nil
}

func SubscribeJSONHelper[T ShellyTRVInfo | ShellyTRVStatus | ShellyDW2Info](
	mqttClient MQTT.Client,
	topic string,
	callback func(T),
) error {
	cb := func(client MQTT.Client, message MQTT.Message) {
		logMessage(message)
		var out T
		err := checkedJSONUnmarshal(message, &out)
		if err != nil {
			return
		}
		callback(out)
	}

	err := checkedSubscribe(mqttClient, topic, cb)
	if err != nil {
		return err
	}
	return nil
}

func SubscribeStringHelper(
	mqttClient MQTT.Client,
	topic string,
	callback func(string),
) error {
	cb := func(client MQTT.Client, message MQTT.Message) {
		logMessage(message)
		callback(string(message.Payload()))
	}

	err := checkedSubscribe(mqttClient, topic, cb)
	if err != nil {
		return err
	}
	return nil
}
