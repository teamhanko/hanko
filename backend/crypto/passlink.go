package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

// PasslinkGenerator will generate a random passlink token
type PasslinkGenerator interface {
	Generate() (string, error)
}

type passlinkGenerator struct {
}

func NewPasslinkGenerator() PasslinkGenerator {
	return &passlinkGenerator{}
}

func (g *passlinkGenerator) Generate() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(bytes), nil
}
