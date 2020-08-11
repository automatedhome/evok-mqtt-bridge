package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/sacOO7/gowebsocket"
	"gopkg.in/yaml.v2"

	types "github.com/automatedhome/evok-mqtt-bridge/pkg/types"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	config     types.Config
	MQTTClient mqtt.Client
	EvokClient gowebsocket.Socket
	send       sync.Mutex
	recv       sync.Mutex
)

func topicMapper(device string, circuit string) string {
	// Map topics to new ones
	for _, m := range config.Mappings {
		if m.Device == device && m.Circuit == circuit {
			return m.Topic
		}
	}
	return "evok/" + device + "/" + circuit + "/value"
}

func applyOffset(input float64, topic string) string {
	offset := 0.0
	for _, m := range config.Mappings {
		if m.Topic == topic {
			offset = m.Offset
			break
		}
	}
	return fmt.Sprintf("%v", input+offset)
}

func onEvokMessage(message string, socket gowebsocket.Socket) {
	var msg types.Message
	if err := json.Unmarshal([]byte(message), &msg); err != nil {
		log.Printf("Failed to unmarshal JSON data from EVOK message: %s\n", message)
		return
	}

	// FIXME: Exclude input 4 as this is constantly floating
	if msg.Device == "input" && msg.Circuit == "4" {
		return
	}

	topic := topicMapper(msg.Device, msg.Circuit)
	v, _ := msg.Value.Float64()
	value := applyOffset(v, topic)

	recv.Lock()
	defer recv.Unlock()
	token := MQTTClient.Publish(topic, 0, false, value)
	token.Wait()
	if token.Error() != nil {
		log.Printf("Failed to publish packet: %s", token.Error())
	}
}

func onMQTTMessage(client mqtt.Client, message mqtt.Message) {
	var msg types.Message
	topic := message.Topic()
	msg.Value = json.Number(message.Payload())
	log.Printf("Received message on MQTT topic: '%s' with payload: '%v'\n", topic, msg.Value)
	msg.Command = "set"
	msg.Device = strings.Split(topic, "/")[1]
	msg.Circuit = strings.Split(topic, "/")[2]

	text, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Wrong data received on MQTT topic '%s' with payload: %+v\n", topic, msg)
		return
	}

	send.Lock()
	defer send.Unlock()
	EvokClient.SendText(string(text))
}

func synchronizer(evok string, interval int) {
	for {
		response, err := http.Get(evok)
		if err != nil {
			log.Fatalf("Couldn't connect to EVOK: %v", err)
		}

		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalf("Couldn't read EVOK data: %v", err)
		}

		data := types.GPIOStates{}
		err = json.Unmarshal([]byte(contents), &data)
		if err != nil {
			log.Printf("Failed to unmarshal JSON data from EVOK message: %v\n", err)
		}

		log.Printf("Got data from evok: %v", data)

		for _, sensor := range data.Data {
			if sensor.Dev != "temp" && sensor.Dev != "relay" && sensor.Dev != "ai" {
				continue
			}
			topic := topicMapper(sensor.Dev, sensor.Circuit)
			value := applyOffset(sensor.Value, topic)
			token := MQTTClient.Publish(topic, 0, false, value)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Failed to publish packet: %s", token.Error())
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func main() {
	broker := flag.String("broker", "tcp://127.0.0.1:1883", "The full url of the MQTT server to connect to ex: tcp://127.0.0.1:1883")
	clientID := flag.String("clientid", "evok", "A clientid for the connection")
	configFile := flag.String("config", "/config.yaml", "Provide configuration file with MQTT topic mappings")
	evok := flag.String("evok", "127.0.0.1:8080", "IP address and port of EVOK API: 127.0.0.1:8080")
	flag.Parse()

	log.Printf("Reading configuration from %s", *configFile)
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("File reading error: %v", err)
		return
	}

	err = yaml.UnmarshalStrict(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Printf("Reading following config from config file: %#v", config)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	opts := mqtt.NewClientOptions().AddBroker(*broker).SetClientID(*clientID)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetAutoReconnect(true)
	opts.OnConnect = func(M mqtt.Client) {
		if token := M.Subscribe("evok/+/+/set", 0, onMQTTMessage); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	MQTTClient = mqtt.NewClient(opts)
	if token := MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Printf("Connected to %s as %s and listening\n", *broker, *clientID)

	EvokClient = gowebsocket.New("ws://" + *evok + "/ws")
	EvokClient.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Println("Recieved connect error ", err)
	}

	EvokClient.OnTextMessage = onEvokMessage

	EvokClient.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from EVOK server ")
		return
	}

	EvokClient.Connect()
	log.Printf("Connected to EVOK on %s\n", *evok)

	go synchronizer("http://"+*evok+"/json/all", config.Interval)

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			EvokClient.Close()
			return
		}
	}
}
