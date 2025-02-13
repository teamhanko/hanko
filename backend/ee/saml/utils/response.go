package utils

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"github.com/beevik/etree"
	rtvalidator "github.com/mattermost/xml-roundtrip-validator"
	"io"
)

const (
	defaultMaxDecompressedResponseSize = 5 * 1024 * 1024
)

func maybeDeflate(data []byte, maxSize int64, decoder func([]byte) error) error {
	err := decoder(data)
	if err == nil {
		return nil
	}

	// Default to 5MB max size
	if maxSize == 0 {
		maxSize = defaultMaxDecompressedResponseSize
	}

	lr := io.LimitReader(flate.NewReader(bytes.NewReader(data)), maxSize+1)

	deflated, err := io.ReadAll(lr)
	if err != nil {
		return err
	}

	if int64(len(deflated)) > maxSize {
		return fmt.Errorf("deflated response exceeds maximum size of %d bytes", maxSize)
	}

	return decoder(deflated)
}

func ParseSamlResponse(samlResponse string) (*etree.Document, *etree.Element, error) {
	raw, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, nil, fmt.Errorf("could not decode saml response: %w", err)
	}

	return parseResponseXml(raw)
}

func parseResponseXml(xml []byte) (*etree.Document, *etree.Element, error) {
	var doc *etree.Document
	var rawXML []byte

	err := maybeDeflate(xml, defaultMaxDecompressedResponseSize, func(xml []byte) error {
		doc = etree.NewDocument()
		rawXML = xml
		return doc.ReadFromBytes(xml)
	})
	if err != nil {
		return nil, nil, err
	}

	el := doc.Root()
	if el == nil {
		return nil, nil, fmt.Errorf("unable to parse response")
	}

	// Examine the response for attempts to exploit weaknesses in Go's encoding/xml
	err = rtvalidator.Validate(bytes.NewReader(rawXML))
	if err != nil {
		return nil, nil, err
	}

	return doc, el, nil
}
