# go-aws-iot-device

[![GoDoc](https://godoc.org/github.com/crow-misia/go-aws-iot-device?status.svg)](https://godoc.org/github.com/crow-misia/go-aws-iot-device)
[![Go Report Card](https://goreportcard.com/badge/github.com/crow-misia/go-aws-iot-device)](https://goreportcard.com/report/github.com/crow-misia/go-aws-iot-device)
[![MIT License](https://img.shields.io/github/license/crow-misia/go-aws-iot-device)](LICENSE)

Package awsiotdevice implements AWS IoT features.

Implemented features:
- Provisioning

# Requirements

Go 1.21 and beyond.

# Install

```shell
go get -u github.com/crow-misia/go-aws-iot-device
```

# Build

```shell
go build
```

### Create certification and key

```shell
aws iot create-keys-and-certificate --set-as-active --certificate-pem-outfile certificate.pem --public-key-outfile public.key --private-key-outfile private.key
```

### endpoint

```shell
aws iot describe-endpoint --endpoint-type iot:Data-ATS
```

# License

MIT
