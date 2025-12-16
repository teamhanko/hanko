package aws_kms

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"sync"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/teamhanko/hanko/backend/v2/config"
)

// AWSKMSAdapter implements crypto.Signer using AWS KMS
type AWSKMSAdapter struct {
	KeyId     string
	KmsClient *kms.Client
	publicKey crypto.PublicKey // cache the public key
	keyMutex  sync.RWMutex     // protect the cached key
}

// NewAWSKMSAdapter initializes and returns a new instance of AWSKMSAdapter with the provided KeyManagement configuration.
func NewAWSKMSAdapter(cfg config.KeyManagement) (*AWSKMSAdapter, error) {
	// LoadDefaultConfig resolves AWS credentials using the following chain (in order of precedence):
	// 1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
	// 2. Shared credentials file (~/.aws/credentials)
	// 3. Shared config file (~/.aws/config)
	// 4. IAM role for Amazon EC2 (via instance metadata service)
	// 5. IAM role for Amazon ECS (via container credentials)
	// 6. IAM role for Amazon EKS (via service account token)
	awsCfg, err := awsConfig.LoadDefaultConfig(
		context.TODO(),
		awsConfig.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	svc := kms.NewFromConfig(awsCfg)

	return &AWSKMSAdapter{
		KeyId:     cfg.KeyID,
		KmsClient: svc,
	}, nil
}

func (k *AWSKMSAdapter) Public() crypto.PublicKey {
	k.keyMutex.RLock()
	if k.publicKey != nil {
		defer k.keyMutex.RUnlock()
		return k.publicKey
	}
	k.keyMutex.RUnlock()

	k.keyMutex.Lock()
	defer k.keyMutex.Unlock()

	// Double-check after acquiring write lock
	if k.publicKey != nil {
		return k.publicKey
	}

	// Fetch and cache the public key
	result, err := k.KmsClient.GetPublicKey(context.TODO(), &kms.GetPublicKeyInput{
		KeyId: &k.KeyId,
	})
	if err != nil {
		log.Printf("failed to get public key: %v", err)
		return nil
	}

	pubKey, err := x509.ParsePKIXPublicKey(result.PublicKey)
	if err != nil {
		log.Printf("failed to parse public key: %v", err)
		return nil
	}

	k.publicKey = pubKey
	return pubKey
}

// Sign implements crypto.Signer interface
func (k *AWSKMSAdapter) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	// Note: rand parameter is not used because AWS KMS handles randomness internally
	_ = rand // explicitly ignore to show intent

	// Determine the signing algorithm based on the hash function
	var signingAlgorithm types.SigningAlgorithmSpec
	switch opts.HashFunc() {
	case crypto.SHA256:
		signingAlgorithm = types.SigningAlgorithmSpecRsassaPkcs1V15Sha256 // or RsassaPkcs1V15Sha256
	case crypto.SHA384:
		signingAlgorithm = types.SigningAlgorithmSpecRsassaPkcs1V15Sha384
	case crypto.SHA512:
		signingAlgorithm = types.SigningAlgorithmSpecRsassaPkcs1V15Sha512
	default:
		return nil, fmt.Errorf("unsupported hash function: %v", opts.HashFunc())
	}

	// Sign the digest using AWS KMS
	result, err := k.KmsClient.Sign(context.TODO(), &kms.SignInput{
		KeyId:            &k.KeyId,
		Message:          digest,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: signingAlgorithm,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign with AWS KMS: %w", err)
	}

	return result.Signature, nil
}

// Verify verifies a signature using the AWS KMS public key
func (k *AWSKMSAdapter) Verify(message, signature []byte, algorithm jwa.SignatureAlgorithm) error {
	// Get the public key
	pubKey := k.Public()
	if pubKey == nil {
		return fmt.Errorf("failed to get public key")
	}

	// Verify the signature directly using crypto operations
	switch algorithm {
	case jwa.RS256:
		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("expected RSA public key for RS256")
		}
		hash := sha256.Sum256(message)
		return rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hash[:], signature)
	default:
		return fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}
