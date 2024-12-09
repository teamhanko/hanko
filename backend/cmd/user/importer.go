package user

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"strings"
	"time"
)

type Importer struct {
	persister       persistence.Persister
	tx              *pop.Connection
	importTimestamp time.Time
}

func (i *Importer) createUser(newUser ImportOrExportEntry) (*models.User, error) {
	userID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	userModel := models.User{
		ID:        userID,
		CreatedAt: i.importTimestamp,
		UpdatedAt: i.importTimestamp,
	}

	if newUser.UserID != "" {
		userModel.ID = uuid.FromStringOrNil(newUser.UserID)
	}

	if newUser.CreatedAt != nil {
		userModel.CreatedAt = newUser.CreatedAt.UTC()
	}

	if newUser.UpdatedAt != nil {
		userModel.UpdatedAt = newUser.UpdatedAt.UTC()
	}

	err = i.persister.GetUserPersisterWithConnection(i.tx).Create(userModel)
	if err != nil {
		return nil, err
	}

	return &userModel, nil
}

func (i *Importer) createEmailAddress(userID uuid.UUID, newEmail ImportOrExportEmail) (*models.Email, error) {
	emailID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	emailModel := models.Email{
		ID:        emailID,
		UserID:    &userID,
		Address:   strings.ToLower(newEmail.Address),
		Verified:  newEmail.IsVerified,
		CreatedAt: i.importTimestamp,
		UpdatedAt: i.importTimestamp,
	}

	err = i.persister.GetEmailPersisterWithConnection(i.tx).Create(emailModel)
	if err != nil {
		return nil, err
	}

	return &emailModel, nil
}

func (i *Importer) createPrimaryEmailAddress(userID uuid.UUID, emailID uuid.UUID) error {
	entryID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	primaryEmailModel := models.PrimaryEmail{
		ID:        entryID,
		EmailID:   emailID,
		UserID:    userID,
		CreatedAt: i.importTimestamp,
		UpdatedAt: i.importTimestamp,
	}

	err = i.persister.GetPrimaryEmailPersisterWithConnection(i.tx).Create(primaryEmailModel)
	return err
}

func (i *Importer) createUsername(userID uuid.UUID, username string) error {
	entryID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	usernameModel := models.Username{
		ID:        entryID,
		UserId:    userID,
		Username:  username,
		CreatedAt: i.importTimestamp,
		UpdatedAt: i.importTimestamp,
	}

	err = i.persister.GetUsernamePersisterWithConnection(i.tx).Create(usernameModel)
	return err
}

func (i *Importer) createWebauthnCredential(userID uuid.UUID, webauthnCredential ImportWebauthnCredential) error {
	createdAt := i.importTimestamp
	updatedAt := i.importTimestamp
	if webauthnCredential.CreatedAt != nil {
		createdAt = webauthnCredential.CreatedAt.UTC()
	}

	if webauthnCredential.UpdatedAt != nil {
		updatedAt = webauthnCredential.UpdatedAt.UTC()
	}

	var transports models.Transports = nil
	for _, transport := range webauthnCredential.Transports {
		transportID, err := uuid.NewV4()
		if err != nil {
			return err
		}
		transports = append(transports, models.WebauthnCredentialTransport{
			ID:                   transportID,
			Name:                 transport,
			WebauthnCredentialID: webauthnCredential.ID,
		})
	}

	webauthnCredentialModel := models.WebauthnCredential{
		ID:              webauthnCredential.ID,
		Name:            webauthnCredential.Name,
		UserId:          userID,
		PublicKey:       webauthnCredential.PublicKey,
		AttestationType: webauthnCredential.AttestationType,
		AAGUID:          webauthnCredential.AAGUID,
		SignCount:       webauthnCredential.SignCount,
		LastUsedAt:      webauthnCredential.LastUsedAt,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		Transports:      transports,
		BackupEligible:  webauthnCredential.BackupEligible,
		BackupState:     webauthnCredential.BackupState,
		MFAOnly:         webauthnCredential.MFAOnly,
	}

	if webauthnCredential.UserHandle != nil {
		existingUserHandle, err := i.persister.GetWebauthnCredentialUserHandlePersisterWithConnection(i.tx).GetByHandle(*webauthnCredential.UserHandle)
		if err != nil {
			return err
		}

		if existingUserHandle != nil {
			webauthnCredentialModel.UserHandleID = &existingUserHandle.ID
		} else {
			userHandleID, err := uuid.NewV4()
			if err != nil {
				return err
			}

			userHandle := models.WebauthnCredentialUserHandle{
				ID:        userHandleID,
				UserID:    userID,
				Handle:    *webauthnCredential.UserHandle,
				CreatedAt: i.importTimestamp,
				UpdatedAt: i.importTimestamp,
			}
			webauthnCredentialModel.UserHandle = &userHandle
			webauthnCredentialModel.UserHandleID = &userHandleID
		}
	}

	err := i.persister.GetWebauthnCredentialPersisterWithConnection(i.tx).Create(webauthnCredentialModel)
	return err
}

func (i *Importer) createPasswordCredential(userID uuid.UUID, passwordCredential ImportPasswordCredential) error {
	passwordID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	createdAt := i.importTimestamp
	updatedAt := i.importTimestamp
	if passwordCredential.CreatedAt != nil {
		createdAt = passwordCredential.CreatedAt.UTC()
	}

	if passwordCredential.UpdatedAt != nil {
		updatedAt = passwordCredential.UpdatedAt.UTC()
	}

	passwordModel := models.PasswordCredential{
		ID:        passwordID,
		UserId:    userID,
		Password:  passwordCredential.Password,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	err = i.persister.GetPasswordCredentialPersisterWithConnection(i.tx).Create(passwordModel)
	return err
}

func (i *Importer) createOTPSecret(userID uuid.UUID, otpSecret ImportOTPSecret) error {
	otpSecretID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	createdAt := i.importTimestamp
	updatedAt := i.importTimestamp
	if otpSecret.CreatedAt != nil {
		createdAt = otpSecret.CreatedAt.UTC()
	}

	if otpSecret.UpdatedAt != nil {
		updatedAt = otpSecret.UpdatedAt.UTC()
	}

	otpSecretModel := models.OTPSecret{
		ID:        otpSecretID,
		UserID:    userID,
		Secret:    otpSecret.Secret,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	err = i.persister.GetOTPSecretPersisterWithConnection(i.tx).Create(otpSecretModel)
	return err
}
