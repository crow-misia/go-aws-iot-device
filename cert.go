package awsiotdevice

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
)

// certificateRequestData 証明書要求データ
type certificateRequestData struct {
	// 秘密鍵
	privateKey string
	// CSR
	csr string
}

func createCsr(curve elliptic.Curve, signatureAlgorithm x509.SignatureAlgorithm) (*certificateRequestData, error) {
	// generate Private Key
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	// create CSR
	csrTemplate := x509.CertificateRequest{
		Subject:            pkix.Name{CommonName: "AWS IoT Certificate"},
		SignatureAlgorithm: signatureAlgorithm,
	}
	csrCertificate, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, privateKey)
	if err != nil {
		return nil, err
	}
	csr := pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csrCertificate,
	})

	// serialize key and CSR
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type: "ECDH PRIVATE KEY", Bytes: privateKeyBytes,
	})

	return &certificateRequestData{
		privateKey: string(privateKeyPem),
		csr:        string(csr),
	}, nil
}
