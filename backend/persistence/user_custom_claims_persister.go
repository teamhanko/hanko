package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// CustomClaimsLimitExceededError is returned when the serialized claims blob exceeds its
// length limit.
type CustomClaimsLimitExceededError struct {
	ValidationErrors *validate.Errors
}

func (e *CustomClaimsLimitExceededError) Error() string {
	return fmt.Sprintf(
		"custom claims limit exceeded: %s",

		strings.Replace(e.ValidationErrors.Error(), "\n", ", ", -1))
}

func (e *CustomClaimsLimitExceededError) Unwrap() error {
	return e.ValidationErrors
}

// IsCustomClaimsLimitExceededError checks if the error is a CustomClaimsLimitExceededError
func IsCustomClaimsLimitExceededError(err error) bool {
	var customClaimsLimitExceededError *CustomClaimsLimitExceededError
	ok := errors.As(err, &customClaimsLimitExceededError)
	return ok
}

type UserCustomClaimsPersister interface {
	Get(userID uuid.UUID, tenantID uuid.UUID) (*models.UserCustomClaims, error)
	Update(customClaims *models.UserCustomClaims) error
}

type userCustomClaimsPersister struct {
	db *pop.Connection
}

func NewUserCustomClaimsPersister(db *pop.Connection) UserCustomClaimsPersister {
	return &userCustomClaimsPersister{db: db}
}

func (p *userCustomClaimsPersister) Get(userID uuid.UUID, tenantID uuid.UUID) (*models.UserCustomClaims, error) {
	customClaims := &models.UserCustomClaims{}
	query := p.db.Where("user_id = ?", userID)
	query = query.Where("tenant_id = ?", tenantID)
	err := query.First(customClaims)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Create an empty row if none exists yet.
			customClaims = &models.UserCustomClaims{UserID: userID, TenantID: tenantID}
			err = p.db.Create(customClaims)
			if err != nil {
				return nil, err
			}
			return customClaims, nil
		}
		return nil, err
	}
	return customClaims, nil
}

func (p *userCustomClaimsPersister) Update(customClaims *models.UserCustomClaims) error {
	vErr, err := p.db.ValidateAndUpdate(customClaims)
	if err != nil {
		return fmt.Errorf("failed to update custom claims: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		limitErrors := validate.NewErrors()
		for key, errs := range vErr.Errors {
			if key == "claims" {
				for _, errMsg := range errs {
					if strings.Contains(errMsg, "custom claims must not exceed") {
						limitErrors.Add(key, errMsg)
					}
				}
			}
		}
		if limitErrors.HasAny() {
			return &CustomClaimsLimitExceededError{ValidationErrors: limitErrors}
		}
		return fmt.Errorf("custom claims validation failed: %w", vErr)
	}

	return nil
}
