// Copyright 2023 Zenichi Amano.

package awsiotdevice

import (
	"crypto/tls"
	autopaho "github.com/eclipse/paho.golang/autopaho"
	"log/slog"
	"os"
)

// ClientOption type
type ClientOption func(*client) error

// WithLogger is logger setter
func WithLogger(logger *slog.Logger) ClientOption {
	return func(client *client) error {
		client.logger = logger
		return nil
	}
}

// WithRootCAFile is Root CA PEM with file setter
func WithRootCAFile(caFile string) ClientOption {
	return func(client *client) error {
		var err error
		client.rootCA, err = os.ReadFile(caFile)
		return err
	}
}

// WithRootCA is Root CA PEM setter
func WithRootCA(pem []byte) ClientOption {
	return func(client *client) error {
		client.rootCA = pem
		return nil
	}
}

// WithCertificateAndPrivateKey is Certificate setter
func WithCertificateAndPrivateKey(certFile, keyFile string) ClientOption {
	return func(client *client) error {
		var err error
		client.certificate, err = tls.LoadX509KeyPair(certFile, keyFile)
		return err
	}
}

// WithCertificate is Certificate setter
func WithCertificate(certificate tls.Certificate) ClientOption {
	return func(client *client) error {
		client.certificate = certificate
		return nil
	}
}

// WithTLSConfig is TLS Config setter
func WithTLSConfig(tlsConfig *tls.Config) ClientOption {
	return func(client *client) error {
		client.tlsConfig = tlsConfig
		return nil
	}
}

// WithClientConfig is MQTT Client config setter
func WithClientConfig(config *autopaho.ClientConfig) ClientOption {
	return func(client *client) error {
		client.clientConfig = config
		return nil
	}
}
