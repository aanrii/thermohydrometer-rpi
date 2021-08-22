package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aanrii/thermohydrometer-rpi/lib"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"golang.org/x/xerrors"
)

type DeviceShadow struct {
	State struct {
		Reported *lib.Metrics `json:"reported"`
	} `json:"state"`
}

func NewDeviceShadow(m *lib.Metrics) *DeviceShadow {
	return &DeviceShadow{
		State: struct {
			Reported *lib.Metrics `json:"reported"`
		}{
			Reported: m,
		},
	}
}

type DeviceShadowClientConfig struct {
	TLS             *tls.Config
	BrokerEndpoint  string
	ClientID        string
	ThingName       string
	ConnectTimeout  time.Duration
	RequestTimeout  time.Duration
	ResponseTimeout time.Duration
}

type DeviceShadowClient struct {
	client mqtt.Client
	conf   *DeviceShadowClientConfig
}

func NewDeviceShadowClient(conf *DeviceShadowClientConfig) (*DeviceShadowClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(conf.BrokerEndpoint)
	opts.SetTLSConfig(conf.TLS)
	opts.SetClientID(conf.ClientID)
	opts.SetCleanSession(false)

	client := mqtt.NewClient(opts)

	t := client.Connect()
	if !t.WaitTimeout(conf.ConnectTimeout) {
		return nil, xerrors.New("Connection timed out")
	}
	if err := t.Error(); err != nil {
		return nil, xerrors.Errorf("Failed to connect to broker: %w", err)
	}

	return &DeviceShadowClient{client, conf}, nil
}

const disconnectTimeoutMillis = 10000

func (c *DeviceShadowClient) Close() {
	c.client.Disconnect(disconnectTimeoutMillis)
}

const qos = 1

func (c *DeviceShadowClient) Upload(s *DeviceShadow) error {
	payload, err := json.Marshal(s)
	if err != nil {
		return xerrors.Errorf("Failed to create payload from given device shadow: %w", err)
	}

	ch := make(chan mqtt.Message, 1)
	mh := func(_ mqtt.Client, m mqtt.Message) { ch <- m }

	acceptedTopicName := fmt.Sprintf("$aws/things/%s/shadow/update/accepted", c.conf.ThingName)
	rejectedTopicName := fmt.Sprintf("$aws/things/%s/shadow/update/rejected", c.conf.ThingName)

	t := c.client.SubscribeMultiple(map[string]byte{
		acceptedTopicName: qos,
		rejectedTopicName: qos,
	}, mh)
	if !t.WaitTimeout(c.conf.RequestTimeout) {
		return xerrors.Errorf("Subscribing to topic %s and %s timed out", acceptedTopicName, rejectedTopicName)
	}
	if err := t.Error(); err != nil {
		return xerrors.Errorf("Failed to subscribe topic %s and %s: %w", acceptedTopicName, rejectedTopicName, err)
	}

	updateTopicName := fmt.Sprintf("$aws/things/%s/shadow/update", c.conf.ThingName)
	t = c.client.Publish(updateTopicName, qos, false, payload)
	if !t.WaitTimeout(c.conf.RequestTimeout) {
		return xerrors.Errorf("Publishing the device shadow to broker timed out")
	}
	if err := t.Error(); err != nil {
		return xerrors.Errorf("Failed to publish the device shadow: %w", err)
	}

	select {
	case m := <-ch:
		switch m.Topic() {
		case acceptedTopicName:
			return nil
		case rejectedTopicName:
			return xerrors.Errorf("The device shadow was rejected: %s", string(m.Payload()))
		default:
			return xerrors.Errorf("Received a message from unknown topic (%s): %s", m.Topic(), string(m.Payload()))
		}
	case <-time.After(c.conf.ResponseTimeout):
		return xerrors.New("Waiting for response timed out")
	}
}
