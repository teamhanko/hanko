package intern

import (
	"encoding/base64"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func WebauthnCredentialToModel(credential *webauthn.Credential, userId uuid.UUID) *models.WebauthnCredential {
	now := time.Now()
	aaguid, _ := uuid.FromBytes(credential.Authenticator.AAGUID)
	return &models.WebauthnCredential{
		ID:              base64.RawURLEncoding.EncodeToString(credential.ID),
		UserId:          userId,
		PublicKey:       base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		AttestationType: credential.AttestationType,
		AAGUID:          aaguid,
		SignCount:       int(credential.Authenticator.SignCount),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func WebauthnCredentialFromModel(credential *models.WebauthnCredential) *webauthn.Credential {
	credId, _ := base64.RawURLEncoding.DecodeString(credential.ID)
	pKey, _ := base64.RawURLEncoding.DecodeString(credential.PublicKey)
	return &webauthn.Credential{
		ID:              credId,
		PublicKey:       pKey,
		AttestationType: credential.AttestationType,
		Authenticator: webauthn.Authenticator{
			AAGUID:    credential.AAGUID.Bytes(),
			SignCount: uint32(credential.SignCount),
		},
	}
}
