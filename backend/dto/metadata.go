package dto

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"strings"

	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

// Metadata represents user metadata with public and unsafe fields
type Metadata struct {
	Public json.RawMessage `json:"public_metadata,omitempty"`
	Unsafe json.RawMessage `json:"unsafe_metadata,omitempty"`
}

// NewMetadata creates a new Metadata DTO from a UserMetadata model
func NewMetadata(metadata *models.UserMetadata) *Metadata {
	if metadata == nil {
		return nil
	}

	result := &Metadata{}

	if metadata.Public.Valid && metadata.Public.String != "{}" {
		result.Public = json.RawMessage(metadata.Public.String)
	}
	if metadata.Unsafe.Valid && metadata.Unsafe.String != "{}" {
		result.Unsafe = json.RawMessage(metadata.Unsafe.String)
	}

	if result.Public == nil && result.Unsafe == nil {
		return nil
	}

	return result
}

// MetadataJWT represents user metadata with public and unsafe fields. This metadata representation is used
// for JWT template processing. Fields are private on purpose since the type provides dedicated methods with the same
// name for accessing the data during template processing.
type MetadataJWT struct {
	public json.RawMessage
	unsafe json.RawMessage
}

// NewMetadataJWT creates a new MetadataJWT from public and unsafe metadata JSON raw messages. Primarily used in tests
// to construct a MetadataJWT (due to private fields)
func NewMetadataJWT(public, unsafe json.RawMessage) *MetadataJWT {
	return &MetadataJWT{
		public: public,
		unsafe: unsafe,
	}
}

// MetadataJWTFromUserModel creates a new MetadataJWT DTO from a UserMetadata model
func MetadataJWTFromUserModel(metadata *models.UserMetadata) *MetadataJWT {
	if metadata == nil {
		return nil
	}

	result := &MetadataJWT{}

	if metadata.Public.Valid && metadata.Public.String != "{}" {
		result.public = json.RawMessage(metadata.Public.String)
	}
	if metadata.Unsafe.Valid && metadata.Unsafe.String != "{}" {
		result.unsafe = json.RawMessage(metadata.Unsafe.String)
	}

	if result.public == nil && result.unsafe == nil {
		return nil
	}

	return result
}

func (m *MetadataJWT) Public(path ...string) string {
	if len(path) < 1 {
		return gjson.GetBytes(m.public, "@this").String()
	}

	return gjson.GetBytes(m.public, strings.Join(path, ".")).String()
}

func (m *MetadataJWT) Unsafe(path ...string) string {
	if len(path) < 1 {
		return gjson.GetBytes(m.unsafe, "@this").String()
	}

	return gjson.GetBytes(m.unsafe, strings.Join(path, ".")).String()
}

func (m *MetadataJWT) String() string {
	if m == nil {
		return ""
	}
	jsonBytes, _ := json.Marshal(m)
	return string(jsonBytes)
}

func (m *MetadataJWT) MarshalJSON() ([]byte, error) {
	s := struct {
		Public json.RawMessage `json:"public_metadata,omitempty"`
		Unsafe json.RawMessage `json:"unsafe_metadata,omitempty"`
	}{
		Public: m.public,
		Unsafe: m.unsafe,
	}

	jsonBytes, err := json.Marshal(s)
	return jsonBytes, err
}
