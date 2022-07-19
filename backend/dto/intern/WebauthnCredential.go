package intern

import (
	"encoding/base64"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func WebauthnCredentialToModel(credential *webauthn.Credential, userId uuid.UUID) *models.WebauthnCredential {
	now := time.Now()
	aaguid, _ := uuid.FromBytes(credential.Authenticator.AAGUID)
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)

	c := &models.WebauthnCredential{
		ID:              credentialID,
		UserId:          userId,
		PublicKey:       base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		AttestationType: credential.AttestationType,
		AAGUID:          aaguid,
		SignCount:       int(credential.Authenticator.SignCount),
		CreatedAt:       now,
		UpdatedAt:       now,
		Transports:      make([]models.WebauthnCredentialTransport, len(credential.Transport)),
	}

	for i, name := range credential.Transport {
		id, _ := uuid.NewV4()
		c.Transports[i] = models.WebauthnCredentialTransport{
			ID:                   id,
			Name:                 string(name),
			WebauthnCredentialID: credentialID,
		}
	}

	return c
}

func WebauthnCredentialFromModel(credential *models.WebauthnCredential) *webauthn.Credential {
	credId, _ := base64.RawURLEncoding.DecodeString(credential.ID)
	pKey, _ := base64.RawURLEncoding.DecodeString(credential.PublicKey)
	transport := make([]protocol.AuthenticatorTransport, len(credential.Transports))

	for i, t := range credential.Transports {
		transport[i] = protocol.AuthenticatorTransport(t.Name)
	}

	return &webauthn.Credential{
		ID:              credId,
		PublicKey:       pKey,
		AttestationType: credential.AttestationType,
		Authenticator: webauthn.Authenticator{
			AAGUID:    credential.AAGUID.Bytes(),
			SignCount: uint32(credential.SignCount),
		},
		Transport: transport,
	}
}
