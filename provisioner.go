package awsiotdevice

import (
	"crypto/elliptic"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"strings"
)

const (
	createKeysAndCertificateTopic = "$aws/certificates/create/json"
	createCertificateFromCsrTopic = "$aws/certificates/create-from-csr/json"
	registerThingTopic            = "$aws/provisioning-templates/%s/provision/json"
)

type Provisioner interface {
	Provisioning(templateName string, parameters map[string]interface{}) (*ProvisioningResponse, error)
	ProvisioningWithCsr(templateName string, parameters map[string]interface{}) (*ProvisioningResponse, error)
}

type ProvisioningResponse struct {
	DeviceConfiguration map[string]interface{}
	ThingName           string
	CertificateId       string
	Certificate         string
	PrivateKey          string
}

type provisioner struct {
	client             Client
	curve              elliptic.Curve
	signatureAlgorithm x509.SignatureAlgorithm
}

func CreateProvisioner(client Client, options ...ProvisionerOption) Provisioner {
	p := &provisioner{
		client:             client,
		curve:              elliptic.P256(),
		signatureAlgorithm: x509.ECDSAWithSHA256,
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// ProvisionerOption type
type ProvisionerOption func(provisioner *provisioner)

// WithCurve is curve implements setter
func WithCurve(curve elliptic.Curve) ProvisionerOption {
	return func(provisioner *provisioner) {
		provisioner.curve = curve
	}
}

// WithSignatureAlgorithm is signature algorithm setter
func WithSignatureAlgorithm(signatureAlgorithm x509.SignatureAlgorithm) ProvisionerOption {
	return func(provisioner *provisioner) {
		provisioner.signatureAlgorithm = signatureAlgorithm
	}
}

func (p *provisioner) Provisioning(templateName string, parameters map[string]interface{}) (*ProvisioningResponse, error) {
	msg, err := p.client.PublishWithReply(createKeysAndCertificateTopic, `{}`)
	if err != nil {
		return nil, err
	}

	var createResponse CreateKeysAndCertificateResponse
	if err = handingReply(msg, &createResponse); err != nil {
		return nil, err
	}

	provisioningResponse, err := p.registerThings(templateName, parameters, createResponse.CertificateOwnershipToken)
	if err != nil {
		return nil, err
	}

	return &ProvisioningResponse{
		DeviceConfiguration: provisioningResponse.DeviceConfiguration,
		ThingName:           provisioningResponse.ThingName,
		PrivateKey:          createResponse.PrivateKey,
		Certificate:         createResponse.CertificatePem,
		CertificateId:       createResponse.CertificateId,
	}, nil
}

func (p *provisioner) ProvisioningWithCsr(templateName string, parameters map[string]interface{}) (*ProvisioningResponse, error) {
	var msg mqtt.Message

	csr, err := createCsr(p.curve, p.signatureAlgorithm)
	if err != nil {
		return nil, err
	}

	request, err := json.Marshal(&CreateCertificateFromCsrRequest{
		CertificateSigningRequest: csr.csr,
	})
	if err != nil {
		return nil, err
	}

	msg, err = p.client.PublishWithReply(createCertificateFromCsrTopic, request)
	if err != nil {
		return nil, err
	}

	var createResponse CreateCertificateFromCsrResponse
	if err = handingReply(msg, &createResponse); err != nil {
		return nil, err
	}

	provisioningResponse, err := p.registerThings(templateName, parameters, createResponse.CertificateOwnershipToken)
	if err != nil {
		return nil, err
	}

	return &ProvisioningResponse{
		DeviceConfiguration: provisioningResponse.DeviceConfiguration,
		ThingName:           provisioningResponse.ThingName,
		PrivateKey:          csr.privateKey,
		Certificate:         createResponse.CertificatePem,
		CertificateId:       createResponse.CertificateId,
	}, nil
}

func (p *provisioner) registerThings(templateName string, parameters map[string]interface{}, token string) (*RegisterThingResponse, error) {
	var msg mqtt.Message

	request, err := json.Marshal(&RegisterThingRequest{
		CertificateOwnershipToken: token,
		Parameters:                parameters,
	})
	if err != nil {
		return nil, err
	}

	msg, err = p.client.PublishWithReply(fmt.Sprintf(registerThingTopic, templateName), request)
	if err != nil {
		return nil, err
	}

	var response RegisterThingResponse
	if err = handingReply(msg, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func handingReply(msg mqtt.Message, response any) error {
	if strings.LastIndex(msg.Topic(), "/rejected") < 0 {
		if err := json.Unmarshal(msg.Payload(), response); err != nil {
			return err
		}
		return nil
	}

	var errorResponse ProvisioningErrorResponse
	if err := json.Unmarshal(msg.Payload(), &errorResponse); err != nil {
		return err
	}

	return errors.New(errorResponse.ErrorMessage)
}
