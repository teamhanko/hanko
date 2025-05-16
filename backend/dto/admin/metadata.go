package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/tidwall/gjson"
	"strings"
)

type PatchMetadataRequest struct {
	Metadata gjson.Result
}

func (m *PatchMetadataRequest) UnmarshalJSON(data []byte) error {
	if !gjson.ValidBytes(data) {
		return fmt.Errorf("body is not valid JSON")
	}

	body := gjson.GetBytes(data, "@this")
	if body.Raw == "null" {
		m.Metadata = body
		return nil
	}

	if body.Raw == "" || (body.Raw != "null" && !body.IsObject()) {
		return errors.New("patch metadata must be null or object")
	}

	var validResultKeys []string
	for _, key := range []string{"public_metadata", "private_metadata", "unsafe_metadata"} {
		prop := gjson.GetBytes(data, key)

		if !prop.Exists() {
			continue
		}

		if prop.Raw != "null" && !prop.IsObject() {
			return fmt.Errorf("%s must be an object or null", key)
		}

		if prop.Raw != "{}" {
			validResultKeys = append(validResultKeys, key)
		}
	}

	m.Metadata = gjson.GetBytes(data, fmt.Sprintf("{%s}", strings.Join(validResultKeys, ",")))

	return nil
}

type Metadata struct {
	Public  json.RawMessage `json:"public_metadata,omitempty"`
	Private json.RawMessage `json:"private_metadata,omitempty"`
	Unsafe  json.RawMessage `json:"unsafe_metadata,omitempty"`
}

func NewMetadata(metadata *models.UserMetadata) *Metadata {
	result := &Metadata{}

	if metadata.Public.Valid && metadata.Public.String != "{}" {
		result.Public = json.RawMessage(metadata.Public.String)
	}
	if metadata.Private.Valid && metadata.Private.String != "{}" {
		result.Private = json.RawMessage(metadata.Private.String)
	}
	if metadata.Unsafe.Valid && metadata.Unsafe.String != "{}" {
		result.Unsafe = json.RawMessage(metadata.Unsafe.String)
	}

	if result.Public == nil && result.Unsafe == nil && result.Private == nil {
		return nil
	}

	return result
}
