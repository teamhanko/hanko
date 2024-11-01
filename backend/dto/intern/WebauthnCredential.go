package intern

import (
	"encoding/base64"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/mapper"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"time"
)

func WebauthnCredentialToModel(credential *webauthn.Credential, userId uuid.UUID, backupEligible, backupState, mfaOnly bool, authenticatorMetadata mapper.AuthenticatorMetadata) *models.WebauthnCredential {
	now := time.Now().UTC()
	aaguid, _ := uuid.FromBytes(credential.Authenticator.AAGUID)
	credentialID := base64.RawURLEncoding.EncodeToString(credential.ID)

	c := &models.WebauthnCredential{
		ID:              credentialID,
		Name:            authenticatorMetadata.GetNameForAaguid(aaguid),
		UserId:          userId,
		PublicKey:       base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		AttestationType: credential.AttestationType,
		AAGUID:          aaguid,
		SignCount:       int(credential.Authenticator.SignCount),
		LastUsedAt:      &now,
		CreatedAt:       now,
		UpdatedAt:       now,
		BackupEligible:  backupEligible,
		BackupState:     backupState,
		MFAOnly:         mfaOnly,
	}

	for _, name := range credential.Transport {
		if string(name) != "" {
			id, _ := uuid.NewV4()
			t := models.WebauthnCredentialTransport{
				ID:                   id,
				Name:                 string(name),
				WebauthnCredentialID: credentialID,
			}
			c.Transports = append(c.Transports, t)
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
