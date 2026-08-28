package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/tidwall/gjson"
)

type CreateUserResponse struct {
	ID      uuid.UUID `json:"id"` // deprecated
	UserID  uuid.UUID `json:"user_id"`
	EmailID uuid.UUID `json:"email_id"`
}

type GetUserResponse struct {
	ID                  uuid.UUID                   `json:"id"`
	Email               *string                     `json:"email,omitempty"`
	Username            *string                     `json:"username,omitempty"`
	WebauthnCredentials []models.WebauthnCredential `json:"webauthn_credentials"` // deprecated
	UpdatedAt           time.Time                   `json:"updated_at"`
	CreatedAt           time.Time                   `json:"created_at"`
	Metadata            *Metadata                   `json:"metadata,omitempty"`
	ProfileData         `json:",inline"`
}

type UserInfoResponse struct {
	ID                    uuid.UUID `json:"id"`
	EmailID               uuid.UUID `json:"email_id"`
	Verified              bool      `json:"verified"`
	HasWebauthnCredential bool      `json:"has_webauthn_credential"`
}

// UserJWT represents an abstracted user model for session management
type UserJWT struct {
	UserID     string       `json:"user_id"`
	Email      *EmailJWT    `json:"email,omitempty"`
	Username   string       `json:"username"`
	Metadata   *MetadataJWT `json:"metadata,omitempty"`
	Name       string       `json:"name"`
	FamilyName string       `json:"family_name"`
	GivenName  string       `json:"given_name"`
	Picture    string       `json:"picture"`

	// customClaims holds this user's tenant-declared custom claims as flat {name: value} JSON,
	// or nil if the user has none. Private, exposed only via the CustomClaims method below (not
	// a field) so JWT templates can call it directly - `.User.CustomClaims "name"`, or bare
	// `.User.CustomClaims` for everything - with no extra hop. Unlike Metadata (which has a
	// public/unsafe split and so needs its own wrapper type with two accessors), there's only
	// one namespace here, so a single method directly on UserJWT is all that's needed.
	customClaims json.RawMessage
}

// CustomClaims returns this user's tenant-declared custom claims: the whole object with no
// argument, or one named claim's value with a single argument - mirroring the calling
// convention of Metadata's Public/Unsafe accessors (path is a gjson path, joined with ".").
// Returns "" for a claim that doesn't exist, or for a user with no custom claims at all.
//
// Returns interface{}, not string, so the value's real JSON type (float64/bool/string/
// []interface{}/map[string]interface{}) survives when Go's text/template evaluates it as part
// of a larger expression - e.g. `{{if .User.CustomClaims "is_staff"}}`: Go's `if`/`and`/`or`/
// `not`/`eq` all inspect the actual reflected value a subexpression produces, which is
// determined by this method's return type, not by how the template's FINAL output gets
// rendered to text. A `string`-typed return would make a `false`-valued boolean claim print as
// the non-empty text "false", which `if` treats as truthy - silently backwards.
//
// This alone does NOT fix a bare `{{ .User.CustomClaims "age" }}` claim template's own output
// type, though: Execute always stringifies whatever the template as a whole renders to,
// regardless of any one method's return type. That's handled separately, by
// session.parseClaimTemplateValue's bare-accessor fast path, which bypasses Execute entirely
// for that one case and calls this same method directly.
func (u *UserJWT) CustomClaims(path ...string) interface{} {
	if u == nil || len(u.customClaims) == 0 {
		return ""
	}
	var result gjson.Result
	if len(path) < 1 {
		result = gjson.GetBytes(u.customClaims, "@this")
	} else {
		result = gjson.GetBytes(u.customClaims, strings.Join(path, "."))
	}
	if !result.Exists() {
		return ""
	}
	return result.Value()
}

func (u *UserJWT) String() string {
	if u == nil {
		return ""
	}

	jsonBytes, _ := json.Marshal(u)
	return string(jsonBytes)
}

// MarshalJSON includes customClaims (unexported, so not covered by the default struct
// marshaling) under the same "custom_claims" key it used to occupy as a field.
func (u *UserJWT) MarshalJSON() ([]byte, error) {
	type userJWTAlias UserJWT
	return json.Marshal(struct {
		*userJWTAlias
		CustomClaims json.RawMessage `json:"custom_claims,omitempty"`
	}{
		userJWTAlias: (*userJWTAlias)(u),
		CustomClaims: u.customClaims,
	})
}

// WithCustomClaims returns a copy of u with its custom claims set to the given flat
// {claimName: value} JSON. Exported setter for the otherwise-private customClaims field, for
// tests in other packages (e.g. session/template_test.go) that construct a UserJWT directly -
// UserJWTFromUserModel (below, same package) sets the field directly instead.
func (u UserJWT) WithCustomClaims(claims json.RawMessage) UserJWT {
	u.customClaims = claims
	return u
}

func UserJWTFromUserModel(userModel *models.User) UserJWT {
	userJWT := UserJWT{
		UserID: userModel.ID.String(),
	}

	if primaryEmail := userModel.Emails.GetPrimary(); primaryEmail != nil {
		userJWT.Email = EmailJWTFromEmailModel(primaryEmail)
	}

	if userModel.Username != nil {
		userJWT.Username = userModel.Username.Username
	}

	if userModel.Metadata != nil {
		metadataJWT := MetadataJWTFromUserModel(userModel.Metadata)
		if metadataJWT != nil {
			userJWT.Metadata = metadataJWT
		}
	}

	if userModel.CustomClaims != nil {
		userJWT.customClaims = CustomClaimsFromUserModel(userModel.CustomClaims)
	}

	if userModel.GivenName.Valid {
		userJWT.GivenName = userModel.GivenName.String
	}

	if userModel.FamilyName.Valid {
		userJWT.FamilyName = userModel.FamilyName.String
	}

	if userModel.Name.Valid {
		userJWT.Name = userModel.Name.String
	}

	if userModel.Picture.Valid {
		userJWT.Picture = userModel.Picture.String
	}

	return userJWT
}
