// Copyright 2023 Zenichi Amano.

package awsiotdevice

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log/slog"
	"time"
)

// Client is AWS IoT Core Client.
type Client interface {
	IsConnected() bool
	IsConnectionOpen() bool
	Connect(clientId string) error
	Disconnect(quiesce uint)
	Publish(topic string, qos byte, retained bool, payload interface{}) error
	PublishWithReply(topic string, payload interface{}) (mqtt.Message, error)
	Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error
	SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) error
	Unsubscribe(topics ...string) error
}

// Client is AWS IoT Core Client.
type client struct {
	logger           *slog.Logger
	endpoint         string
	rootCA           []byte
	certificate      tls.Certificate
	tlsConfig        *tls.Config
	mqttConfig       *mqtt.ClientOptions
	mqttClient       mqtt.Client
	connectTimeout   time.Duration
	publishTimeout   time.Duration
	subscribeTimeout time.Duration
}

// New returns a new AWS IoT Core Client instance.
func New(endpoint string, options ...ClientOption) (Client, error) {
	var err error

	c := &client{
		endpoint:         endpoint,
		connectTimeout:   15 * time.Second,
		publishTimeout:   10 * time.Second,
		subscribeTimeout: 10 * time.Second,
	}

	for _, option := range options {
		if err = option(c); err != nil {
			return nil, err
		}
	}

	if c.tlsConfig == nil {
		if c.tlsConfig, err = newTLSConfig(c.rootCA, c.certificate); err != nil {
			return nil, err
		}
	}

	if c.mqttConfig == nil {
		c.mqttConfig = mqtt.NewClientOptions()
	}

	return c, nil
}

func (c *client) Connect(clientId string) error {
	opts := c.mqttConfig
	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", c.endpoint, 443))
	opts.SetTLSConfig(c.tlsConfig)
	opts.SetClientID(clientId)

	client := mqtt.NewClient(opts)
	c.mqttClient = client

	token := client.Connect()
	return mqtt.WaitTokenTimeout(token, c.connectTimeout)
}

func (c *client) Disconnect(quiesce uint) {
	mqttClient := c.mqttClient
	if mqttClient != nil {
		mqttClient.Disconnect(quiesce)
	}
}

func (c *client) IsConnected() bool {
	return c.mqttClient != nil && c.mqttClient.IsConnected()
}

func (c *client) IsConnectionOpen() bool {
	return c.mqttClient != nil && c.mqttClient.IsConnectionOpen()
}

func (c *client) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := c.mqttClient.Publish(topic, qos, retained, payload)
	return waitTokenTimeout(token, c.publishTimeout)
}

func (c *client) PublishWithReply(topic string, payload interface{}) (mqtt.Message, error) {
	replayCh := make(chan mqtt.Message)
	defer close(replayCh)
	subscribeCallback := func(c mqtt.Client, m mqtt.Message) {
		replayCh <- m
	}
	acceptedTopic := fmt.Sprintf("%s/accepted", topic)
	rejectedTopic := fmt.Sprintf("%s/rejected", topic)
	filters := make(map[string]byte)
	filters[acceptedTopic] = 1
	filters[rejectedTopic] = 1
	subscribeToken := c.mqttClient.SubscribeMultiple(filters, subscribeCallback)
	defer c.mqttClient.Unsubscribe(acceptedTopic, rejectedTopic)
	if err := waitTokenTimeout(subscribeToken, c.subscribeTimeout); err != nil {
		return nil, err
	}
	token := c.mqttClient.Publish(topic, 1, false, payload)
	if err := waitTokenTimeout(token, c.publishTimeout); err != nil {
		return nil, err
	}

	am := <-replayCh
	return am, token.Error()
}

func (c *client) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := c.mqttClient.Subscribe(topic, qos, callback)
	return waitTokenTimeout(token, c.subscribeTimeout)
}

func (c *client) SubscribeMultiple(filters map[string]byte, callback mqtt.MessageHandler) error {
	token := c.mqttClient.SubscribeMultiple(filters, callback)
	return waitTokenTimeout(token, c.subscribeTimeout)
}

func (c *client) Unsubscribe(topics ...string) error {
	token := c.mqttClient.Unsubscribe(topics...)
	return waitTokenTimeout(token, c.subscribeTimeout)
}

func newTLSConfig(rootCA []byte, certificate tls.Certificate) (*tls.Config, error) {
	var err error
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(rootCA)
	if certificate.Leaf == nil {
		if certificate.Leaf, err = x509.ParseCertificate(certificate.Certificate[0]); err != nil {
			return nil, err
		}
	}
	return &tls.Config{
		RootCAs:            certpool,
		Certificates:       []tls.Certificate{certificate},
		NextProtos:         []string{"x-amzn-mqtt-ca"},
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}, nil
}

func waitTokenTimeout(token mqtt.Token, d time.Duration) error {
	if d > 0 {
		return mqtt.WaitTokenTimeout(token, d)
	}
	return token.Error()
}
