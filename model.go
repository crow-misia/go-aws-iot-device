// Copyright 2023 Zenichi Amano.

package awsiotdevice

type CreateKeysAndCertificateResponse struct {
	CertificateId             string `json:"certificateId"`
	CertificatePem            string `json:"certificatePem"`
	PrivateKey                string `json:"privateKey"`
	CertificateOwnershipToken string `json:"certificateOwnershipToken"`
}

type CreateCertificateFromCsrRequest struct {
	CertificateSigningRequest string `json:"certificateSigningRequest"`
}

type CreateCertificateFromCsrResponse struct {
	CertificateId             string `json:"certificateId"`
	CertificatePem            string `json:"certificatePem"`
	CertificateOwnershipToken string `json:"certificateOwnershipToken"`
}

type RegisterThingRequest struct {
	CertificateOwnershipToken string         `json:"certificateOwnershipToken"`
	Parameters                map[string]any `json:"parameters"`
}

type RegisterThingResponse struct {
	DeviceConfiguration map[string]any `json:"deviceConfiguration"`
	ThingName           string         `json:"thingName"`
}

type ProvisioningErrorResponse struct {
	StatusCode   int    `json:"statusCode"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}
