package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/crow-misia/go-aws-iot-device"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"log"
	"os"
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
	client, err := awsiotdevice.New(endpoint,
		awsiotdevice.WithRootCAFile(caFilename),
		awsiotdevice.WithCertificateAndPrivateKey(certFilename, keyFilename),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to construct tls config: %v", err))
	}
	defer client.Disconnect(ctx)

	clientId, err := uuid.NewRandom()
	if err != nil {
		panic(fmt.Sprintf("failed to generate client id: %v", err))
	}
	log.Printf("connecting... %s with %s\n", endpoint, clientId)
	if err = client.Connect(ctx, clientId.String()); err != nil {
		panic(fmt.Sprintf("failed to connect broker: %v", err))
	}

	provisioner := awsiotdevice.CreateProvisioner(client)
	response, err := provisioner.ProvisioningWithCsr(ctx, template, map[string]interface{}{
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
	fmt.Printf("%s\n", jsonStr)
}
