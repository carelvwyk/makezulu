package aws_iot

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const debug = false

// Message represents an incoming message somfr the IoT endpoint
type Message struct {
	Topic   string
	Payload []byte
}

func (m Message) String() string {
	return fmt.Sprintf("TOPIC: %s PAYLOAD: %s", m.Topic, string(m.Payload))
}

// Thing represents an AWS IoT "Thing"
type Thing struct {
	name      string
	tlsConfig *tls.Config
	region    string
	// incoming messages from the IoT endpoint
	incoming chan Message
	// outgoing messages to the IoT endpoint
	outgoing chan interface{}
	quit     chan bool
	running  bool
	client   mqtt.Client
}

// New creates a new reference to an AWS IoT thing in the given AWS region based
// on the provided thing-name and credentials. It does not connect to the
// AWS IoT PubSub endpoint yet, use "Connect" (blocking) to start handling
// sending and receiving of messages.
func New(thingName, privKey, cert, region string) (*Thing, error) {
	t := Thing{name: thingName, region: region}
	t.outgoing = make(chan interface{})
	t.incoming = make(chan Message)
	t.quit = make(chan bool)

	creds, err := tls.X509KeyPair([]byte(cert), []byte(privKey))
	if err != nil {
		return nil, err
	}
	t.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{creds},
	}
	t.tlsConfig.BuildNameToCertificate()

	return &t, nil
}

// Connect connects to the AWS IoT thing and starts handling sending and
// receiving of messages.
func (t *Thing) Connect() error {
	serverURL := fmt.Sprintf("ssl://data.iot.%s.amazonaws.com:8883", t.region)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverURL)
	opts.SetClientID(t.name).SetTLSConfig(t.tlsConfig)
	// Note: You may want to add different handlers for different topics
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		newMsg := Message{Topic: msg.Topic(), Payload: msg.Payload()}
		select {
		case t.incoming <- newMsg:
		default:
			log.Printf("Received message but nobody is listening on this side: %s",
				newMsg.String())
		}
	})

	if debug {
		mqtt.DEBUG = log.New(os.Stdout, "logger: ", log.Lshortfile)
	}

	t.client = mqtt.NewClient(opts)
	if token := t.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer t.client.Disconnect(100)

	log.Printf("%s connected to %s and listening", t.name, serverURL)

	// subscribe to the desired state topic for this thing
	if token := t.client.Subscribe(fmt.Sprintf("$aws/things/%s/shadow/#", t.name),
		0, nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	t.running = true
	// Wait for the stop signal:
	for {
		select {
		case <-t.quit:
			t.running = false
			return nil
		case msg := <-t.outgoing:
			b, err := json.Marshal(&msg)
			if err != nil {
				return err
			}
			if token := t.client.Publish(fmt.Sprintf("$aws/things/%s/shadow/update", t.name),
				0, false, b); token.Wait() && token.Error() != nil {
				return token.Error()
			}
		}
	}
}

// Stop disconnects the AWS IoT client
func (t *Thing) Stop() error {
	select {
	case t.quit <- true:
	default:
		return errors.New("Not running")
	}
	close(t.quit)
	close(t.incoming)
	close(t.outgoing)
	return nil
}

func (t *Thing) SubChannel() <-chan Message {
	return t.incoming
}

func (t *Thing) PubChannel() chan<- interface{} {
	return t.outgoing
}
