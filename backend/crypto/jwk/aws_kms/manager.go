package aws_kms

import (
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v2/config"
)

// AWSKMSManager implements the KeyManager interface using AWS KMS
type AWSKMSManager struct {
	awsAdapter *AWSKMSAdapter
	algorithm  jwa.SignatureAlgorithm
}

func NewAWSKMSManager(cfg config.KeyManagement) (*AWSKMSManager, error) {
	adapter, err := NewAWSKMSAdapter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS KMS adapter: %w", err)
	}

	return &AWSKMSManager{
		awsAdapter: adapter,
		algorithm:  jwa.RS256,
	}, nil
}

func (m *AWSKMSManager) GenerateKey() (jwk.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *AWSKMSManager) GetPublicKeys() (jwk.Set, error) {
	publicKey := m.awsAdapter.Public()
	key, err := jwk.PublicKeyOf(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key: %w", err)
	}

	if err := key.Set(jwk.KeyIDKey, m.awsAdapter.KeyId); err != nil {
		return nil, fmt.Errorf("failed to set key id: %w", err)
	}

	if err := key.Set(jwk.KeyUsageKey, jwk.ForSignature); err != nil {
		return nil, fmt.Errorf("failed to set key usage: %w", err)
	}

	if err := key.Set(jwk.AlgorithmKey, m.algorithm); err != nil {
		return nil, fmt.Errorf("failed to set algorithm: %w", err)
	}

	set := jwk.NewSet()
	if err := set.AddKey(key); err != nil {
		return nil, fmt.Errorf("failed to add key to set: %w", err)
	}

	return set, nil
}

func (m *AWSKMSManager) GetSigningKey() (jwk.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *AWSKMSManager) Sign(token jwt.Token) ([]byte, error) {
	headers := jws.NewHeaders()
	_ = headers.Set(jws.KeyIDKey, m.awsAdapter.KeyId)

	// Use the adapter struct for signing
	signed, err := jwt.Sign(token, jwt.WithKey(m.algorithm, m.awsAdapter, jws.WithProtectedHeaders(headers)))
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT with AWS KMS: %w", err)
	}

	return signed, nil
}

func (m *AWSKMSManager) Verify(bytes []byte) (jwt.Token, error) {
	// Use the adapter struct for verification
	token, err := jwt.Parse(bytes, jwt.WithKey(m.algorithm, m.awsAdapter))
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT with AWS KMS: %w", err)
	}

	return token, nil
}
