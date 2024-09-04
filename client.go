// Copyright 2023 Zenichi Amano.

package awsiotdevice

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"golang.org/x/net/context"
	"log/slog"
	"net/url"
)

// Client is AWS IoT Core Client.
type Client interface {
	Connect(ctx context.Context, clientId string) error
	Disconnect(ctx context.Context)
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

	logger       *slog.Logger
	rootCA       []byte
	certificate  tls.Certificate
	tlsConfig    *tls.Config
	clientConfig *autopaho.ClientConfig
}

// New returns a new AWS IoT Core Client instance.
func New(endpoint string, options ...ClientOption) (Client, error) {
	var err error

	c := &client{}

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

	if c.clientConfig == nil {
		c.clientConfig = &autopaho.ClientConfig{}
	}

	u, err := url.Parse(fmt.Sprintf("ssl://%s:%d", endpoint, 443))
	if err != nil {
		return nil, err
	}
	c.clientConfig.ServerUrls = append(c.clientConfig.ServerUrls, u)

	return c, nil
}

func (c *client) Connect(ctx context.Context, clientId string) error {
	c.clientConfig.ClientID = clientId
	c.clientConfig.TlsCfg = c.tlsConfig

	client, err := autopaho.NewConnection(ctx, *c.clientConfig)
	if err != nil {
		return err
	}
	c.ConnectionManager = client
	return nil
}

func (c *client) Disconnect(ctx context.Context) {
	cm := c.ConnectionManager
	if cm != nil {
		_ = cm.Disconnect(ctx)
		c.ConnectionManager = nil
	}
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
