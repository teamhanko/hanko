package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

// MetadataLimitExceededError is returned when metadata fields exceed their length limits
type MetadataLimitExceededError struct {
	ValidationErrors *validate.Errors
}

func (e *MetadataLimitExceededError) Error() string {
	return fmt.Sprintf(
		"metadata limit exceeded: %s",
		strings.Replace(e.ValidationErrors.Error(), "\n", ", ", -1))
}

func (e *MetadataLimitExceededError) Unwrap() error {
	return e.ValidationErrors
}

// IsMetadataLimitExceededError checks if the error is a MetadataLimitExceededError
func IsMetadataLimitExceededError(err error) bool {
	var metadataLimitExceededError *MetadataLimitExceededError
	ok := errors.As(err, &metadataLimitExceededError)
	return ok
}

type UserMetadataPersister interface {
	Get(userID uuid.UUID) (*models.UserMetadata, error)
	Update(metadata *models.UserMetadata) error
	Delete(metadata *models.UserMetadata) error
}

type userMetadataPersister struct {
	db *pop.Connection
}

func NewUserMetadataPersister(db *pop.Connection) UserMetadataPersister {
	return &userMetadataPersister{db: db}
}

func (p *userMetadataPersister) Get(userID uuid.UUID) (*models.UserMetadata, error) {
	metadata := &models.UserMetadata{}
	err := p.db.Where("user_id = ?", userID).First(metadata)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create new metadata if none exists
			metadata = &models.UserMetadata{UserID: userID}
			err = p.db.Create(metadata)
			if err != nil {
				return nil, err
			}
			return metadata, nil
		}
		return nil, err
	}
	return metadata, nil
}

func (p *userMetadataPersister) Update(metadata *models.UserMetadata) error {
	vErr, err := p.db.ValidateAndUpdate(metadata)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		metadataRangeErrors := validate.NewErrors()
		for key, errs := range vErr.Errors {
			if key == "public" || key == "private" || key == "unsafe" {
				for _, errMsg := range errs {
					if strings.Contains(errMsg, "metadata must not exceed") {
						metadataRangeErrors.Add(key, errMsg)
					}
				}
			}
		}
		if metadataRangeErrors.HasAny() {
			return &MetadataLimitExceededError{ValidationErrors: metadataRangeErrors}
		}
		return fmt.Errorf("metadata validation failed: %w", vErr)
	}

	return nil
}

func (p *userMetadataPersister) Delete(metadata *models.UserMetadata) error {
	err := p.db.Destroy(metadata)
	if err != nil {
		return fmt.Errorf("failed to delete user metadata: %w", err)
	}

	return nil
}
