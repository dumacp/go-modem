package nmea

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-modem/appliance/crosscutting/logs"
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	clietnName    = "go-nmea-actor"
	topicGPS      = "GPS"
	topicEventGPS = "EVENTS/gps"
	topicBadGPS   = "EVENTS/badgps"
)

type actorpubsub struct {
	clientMqtt MQTT.Client
	debug      bool
}

func NewPubSubActor(debug bool) actor.Actor {
	act := &actorpubsub{}
	act.debug = debug
	return act
}

func (act *actorpubsub) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("actor started \"%s\"", ctx.Self().Id)
		clientMqtt, err := connectMqtt()
		if err != nil {
			panic(err)
		}
		act.clientMqtt = clientMqtt
	case *msgGPS:
		token := act.clientMqtt.Publish(topicGPS, 0, false, msg.data)
		if ok := token.WaitTimeout(10 * time.Second); !ok {
			act.clientMqtt.Disconnect(100)
			panic("MQTT connection failed")
		}
	case *msgBadGPS:
		token := act.clientMqtt.Publish(topicBadGPS, 0, false, msg.data)
		if ok := token.WaitTimeout(10 * time.Second); !ok {
			act.clientMqtt.Disconnect(100)
			panic("MQTT connection failed")
		}
	case *eventGPS:
		// fmt.Printf("event: %s\n", msg.event)
		token := act.clientMqtt.Publish(topicEventGPS, 0, false, msg.event)
		if ok := token.WaitTimeout(10 * time.Second); !ok {
			act.clientMqtt.Disconnect(100)
			panic("MQTT connection failed")
		}
	}
}

func connectMqtt() (MQTT.Client, error) {
	opts := MQTT.NewClientOptions().AddBroker("tcp://127.0.0.1:1883")
	opts.SetClientID(clietnName)
	opts.SetAutoReconnect(true)
	conn := MQTT.NewClient(opts)
	token := conn.Connect()
	if ok := token.WaitTimeout(30 * time.Second); !ok {
		return nil, fmt.Errorf("MQTT connection failed")
	}
	return conn, nil
}
