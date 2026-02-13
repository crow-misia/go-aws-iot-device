package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"time"

	awsiotdevice "github.com/crow-misia/go-aws-iot-device"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

func main() {
	var (
		endpoint     string
		caFilename   string
		certFilename string
		keyFilename  string
		template     string
	)
	flag.NewFlagSet("help", flag.ExitOnError)
	flag.StringVar(&endpoint, "endpoint", "", "AWS Endpoint hostname")
	flag.StringVar(&caFilename, "ca", "AmazonRootCA1.pem", "CA Certification PEM filename")
	flag.StringVar(&certFilename, "cert", "certificate.pem", "Client Certification PEM filename")
	flag.StringVar(&keyFilename, "key", "private.key", "Private Key filename")
	flag.StringVar(&template, "template", "", "Template name")
	flag.Parse()

	if len(endpoint) == 0 || len(caFilename) == 0 || len(certFilename) == 0 || len(keyFilename) == 0 || len(template) == 0 {
		flag.PrintDefaults()
		return
	}

	ctx := context.Background()
	log := slog.Default()

	clientId, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("failed to generate client id: %v", err))
	}
	serverUrl, err := url.Parse(fmt.Sprintf("ssl://%s:443", endpoint))
	if err != nil {
		panic(fmt.Sprintf("failed to parse ServerURL(%s): %v", endpoint, err))
	}
	rootCa, err := os.ReadFile(caFilename)
	if err != nil {
		panic(fmt.Sprintf("failed to read Root CA: %v", err))
	}
	cert, err := tls.LoadX509KeyPair(certFilename, keyFilename)
	if err != nil {
		panic(fmt.Sprintf("failed to read Certificate: %v", err))
	}
	tlsCfg, err := awsiotdevice.NewTLSConfig(rootCa, cert)
	if err != nil {
		panic(fmt.Sprintf("failed to parse Certificate: %v", err))
	}

	cfg := autopaho.ClientConfig{
		Debug:      awsiotdevice.NewSlogLogger(ctx, log, slog.LevelInfo),
		ServerUrls: []*url.URL{serverUrl},
		TlsCfg:     tlsCfg,
		ClientConfig: paho.ClientConfig{
			ClientID: clientId.String(),
		},
	}
	log.Info("connecting...", slog.String("endpoint", endpoint), slog.Any("clientId", clientId))
	client, err := awsiotdevice.NewConnection(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect: %v", err))
	}
	defer client.Disconnect(ctx)

	connCtx, cancelFunc := context.WithTimeout(ctx, 20*time.Second)
	defer cancelFunc()
	if err = client.AwaitConnection(connCtx); err != nil {
		panic(fmt.Sprintf("connection error: %v", err))
	}

	provisioner := awsiotdevice.CreateProvisioner(client)
	response, err := provisioner.ProvisioningWithCsr(ctx, template, map[string]any{
		"SerialNumber": clientId.String(),
	})
	if err != nil {
		panic(fmt.Sprintf("failed create CSR: %v", err))
	}

	certificateId := response.CertificateId
	if err = os.WriteFile(fmt.Sprintf("certificate-%s.pem", certificateId), []byte(response.Certificate), 0644); err != nil {
		panic(fmt.Sprintf("failed output certrificate PEM: %v", err))
	}
	if err = os.WriteFile(fmt.Sprintf("private-%s.key", certificateId), []byte(response.PrivateKey), 0644); err != nil {
		panic(fmt.Sprintf("failed output private key: %v", err))
	}
	jsonStr, _ := json.Marshal(response)
	log.Info(string(jsonStr))
}
