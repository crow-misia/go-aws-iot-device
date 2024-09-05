// Copyright 2023 Zenichi Amano.

package awsiotdevice

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"golang.org/x/net/context"
)

// Client is AWS IoT Core Client.
type Client interface {
	Disconnect(ctx context.Context) error
	Done() <-chan struct{}
	AwaitConnection(ctx context.Context) error
	Authenticate(ctx context.Context, a *paho.Auth) (*paho.AuthResponse, error)
	Subscribe(ctx context.Context, s *paho.Subscribe) (*paho.Suback, error)
	Unsubscribe(ctx context.Context, u *paho.Unsubscribe) (*paho.Unsuback, error)
	Publish(ctx context.Context, p *paho.Publish) (*paho.PublishResponse, error)
	PublishWithReply(ctx context.Context, p *paho.Publish) (*paho.Publish, error)
	PublishViaQueue(ctx context.Context, p *autopaho.QueuePublish) error
	AddOnPublishReceived(f func(autopaho.PublishReceived) (bool, error)) func()
}

// Client is AWS IoT Core Client.
type client struct {
	*autopaho.ConnectionManager
}

// NewConnection returns a new AWS IoT Core Client instance.
func NewConnection(ctx context.Context, config autopaho.ClientConfig) (Client, error) {
	cm, err := autopaho.NewConnection(ctx, config)
	if err != nil {
		return nil, err
	}
	return &client{
		ConnectionManager: cm,
	}, nil
}

func (c *client) Disconnect(ctx context.Context) error {
	var err error
	cm := c.ConnectionManager
	if cm != nil {
		err = cm.Disconnect(ctx)
		c.ConnectionManager = nil
	}
	return err
}

func (c *client) PublishWithReply(ctx context.Context, p *paho.Publish) (*paho.Publish, error) {
	acceptedTopic := fmt.Sprintf("%s/accepted", p.Topic)
	rejectedTopic := fmt.Sprintf("%s/rejected", p.Topic)

	s := &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{Topic: acceptedTopic, QoS: 1},
			{Topic: rejectedTopic, QoS: 1},
		},
	}
	u := &paho.Unsubscribe{
		Topics: []string{acceptedTopic, rejectedTopic},
	}

	replayCh := make(chan *paho.Publish)
	defer close(replayCh)

	unsubscribeFn := c.AddOnPublishReceived(func(received autopaho.PublishReceived) (bool, error) {
		if received.AlreadyHandled {
			return false, nil
		}
		replayCh <- received.Packet
		return true, nil
	})
	defer unsubscribeFn()

	_, err := c.Subscribe(ctx, s)
	if err != nil {
		return nil, err
	}
	defer c.Unsubscribe(ctx, u)

	_, err = c.Publish(ctx, p)
	if err != nil {
		return nil, err
	}

	am := <-replayCh
	return am, nil
}

func NewTLSConfig(rootCA []byte, certificate tls.Certificate) (*tls.Config, error) {
	var err error
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(rootCA)
	if certificate.Leaf == nil {
		if certificate.Leaf, err = x509.ParseCertificate(certificate.Certificate[0]); err != nil {
			return nil, err
		}
	}
	return &tls.Config{
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{certificate},
		NextProtos:         []string{"x-amzn-mqtt-ca"},
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
	}, nil
}
