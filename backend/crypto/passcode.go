package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

type PasscodeGenerator interface {
	Generate() (string, error)
}

type passcodeGenerator struct {
}

func NewPasscodeGenerator() PasscodeGenerator {
	return &passcodeGenerator{}
}

func (g *passcodeGenerator) Generate() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}
	return fmt.Sprintf("%06d", n), nil
}
