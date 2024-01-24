package mapper

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/knadh/koanf/parsers/json"
	"github.com/teamhanko/hanko/backend/config"
	"log"
)

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
	k, err := config.LoadFile(authMetaFilePath, json.Parser())

	if err != nil {
		log.Println(err)
		return nil
	}

	if k == nil {
		log.Println("no authenticator metadata file provided. Skipping...")
		return nil
	}

	var authenticatorMetadata AuthenticatorMetadata
	err = k.Unmarshal("", &authenticatorMetadata)
	if err != nil {
		log.Println(fmt.Errorf("unable to unmarshal authenticator metadata: %w", err))
		return nil
	}

	return authenticatorMetadata
}
