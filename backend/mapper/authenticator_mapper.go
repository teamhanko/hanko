package mapper

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	kjson "github.com/knadh/koanf/parsers/json"
	"github.com/teamhanko/hanko/backend/config"
	"log"
)

//go:embed aaguid.json
var authenticatorMetadataJson []byte

type Authenticator struct {
	Name      string `json:"name"`
	IconLight string `json:"icon_light"`
	IconDark  string `json:"icon_dark"`
}

type AuthenticatorMetadata map[string]Authenticator

func (am AuthenticatorMetadata) GetNameForAaguid(aaguid uuid.UUID) *string {
	if am != nil {
		if authenticatorMetadata, ok := am[aaguid.String()]; ok {
			return &authenticatorMetadata.Name
		}
	}

	return nil
}

func LoadAuthenticatorMetadata(authMetaFilePath *string) AuthenticatorMetadata {
	k, err := config.LoadFile(authMetaFilePath, kjson.Parser())

	if err != nil {
		log.Println(err)
		return nil
	}

	var authenticatorMetadata AuthenticatorMetadata

	if k == nil {
		log.Println("no authenticator metadata file provided. Using embedded one.")
		authenticatorMetadata, err = loadFromEmbeddedFile()
		if err != nil {
			log.Println("no valid authenticator metadata file provided. Skipping...")
		}

		return authenticatorMetadata
	}

	err = k.Unmarshal("", &authenticatorMetadata)
	if err != nil {
		log.Println(fmt.Errorf("unable to unmarshal authenticator metadata: %w", err))
		return nil
	}

	return authenticatorMetadata
}

func loadFromEmbeddedFile() (AuthenticatorMetadata, error) {
	var authMeta AuthenticatorMetadata
	err := json.Unmarshal(authenticatorMetadataJson, &authMeta)
	if err != nil {
		log.Println(fmt.Errorf("unable to unmarshal authenticator metadata: %w", err))
		return nil, err
	}

	return authMeta, nil
}
