package jwk

import (
	"fmt"

	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk/aws_kms"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

func NewManager(cfg config.Config, persister persistence.Persister, multitenancy bool) (KeyProvider, error) {
	switch cfg.Secrets.KeyManagement.Type {
	case "local":
		return local_db.NewDefaultManager(cfg.SecretKeys, persister.GetJwkPersister(), multitenancy)
	case "aws_kms":
		return aws_kms.NewAWSKMSManager(cfg.Secrets.KeyManagement, multitenancy)
	}

	return nil, fmt.Errorf("unsupported key management type: %s", cfg.Secrets.KeyManagement.Type)
}
