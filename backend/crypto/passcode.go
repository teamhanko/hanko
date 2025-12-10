package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type PasscodeGenerator interface {
	Generate() (string, error)
}

type numericPasscodeGenerator struct {
}

func NewNumericPasscodeGenerator() PasscodeGenerator {
	return &numericPasscodeGenerator{}
}

func (g *numericPasscodeGenerator) Generate() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return fmt.Sprintf("%06d", n), nil
}

type alphanumericPasscodeGenerator struct {
}

// alphanumericChars without ambiguous characters: 0 (zero), 1 (one), O (oh), I, (eye), l (ell)
const alphanumericChars = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func NewAlphanumericPasscodeGenerator() PasscodeGenerator {
	return &alphanumericPasscodeGenerator{}
}

func (a *alphanumericPasscodeGenerator) Generate() (string, error) {
	b := make([]byte, 6)
	max := big.NewInt(int64(len(alphanumericChars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		b[i] = alphanumericChars[n.Int64()]
	}
	return string(b), nil
}
