package security

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type extractSignatureTestSuite struct {
	suite.Suite
}

func (suite *extractSignatureTestSuite) TestExtractSignatureHeaders() {
	signature := `keyId="https://pixel.netzspielplatz.de/i/actor#main-key",headers="(request-target) date host accept content-type user-agent",algorithm="rsa-sha256",signature="NkKmcSp2G10sBs7eMAB0OgOeqExSvkbuRZAiIVLtGA9NJdnRpf3JEdTGyTKEHs7ywKoz7xEF9AUmJGfsFx+IpbIOyWfjXHAukcLX3UA/dK64gqR5x0VKOisf+wmNb2KixpZ8dJbFhAc6i4Y85gRATVuVje17KZloNTS0rkb30U/g3fQjfRmegcEPF5P+srH91rAREMP0kMEjlHf4IyZh1+6l/yDfrKUakK7auP4tF1Obhxf7XoWB2ouq+H8scz4MYHKg1jQeASwzFlj+5osWLgoBajRc0gGBSQ7mCvbeKw5hKyl3XJ9f3JAKTxyg8UciZ//jY7Ejr77ncesnl/zwMQ=="`
	signatureHeaders := extractSignatureHeaders(signature)
	suite.EqualValues([]string{
		"(request-target)",
		"date",
		"host",
		"accept",
		"content-type",
		"user-agent",
	}, signatureHeaders)
}

func TestExtractSignatureTestSuite(t *testing.T) {
	suite.Run(t, &extractSignatureTestSuite{})
}
