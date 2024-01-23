package mapper

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/knadh/koanf/parsers/json"
	"github.com/teamhanko/hanko/backend/config"
	"log"
)

type Aaguid struct {
	Name      string `json:"name"`
	IconLight string `json:"icon_light"`
	IconDark  string `json:"icon_dark"`
}

type AaguidMap map[string]Aaguid

func (w AaguidMap) GetNameForAaguid(aaguid uuid.UUID) *string {
	if webauthnAaguid, ok := w[aaguid.String()]; ok {
		return &webauthnAaguid.Name
	} else {
		return nil
	}
}

func LoadAaguidMap(aaguidFilePath *string) AaguidMap {
	k, err := config.LoadFile(aaguidFilePath, json.Parser())

	if err != nil {
		log.Println(err)
		return nil
	}

	if k == nil {
		log.Println("no aaguid map file provided. Skipping...")
		return nil
	}

	var aaguidMap AaguidMap
	err = k.Unmarshal("", &aaguidMap)
	if err != nil {
		log.Println(fmt.Errorf("unable to unmarshal aaguid map: %w", err))
		return nil
	}

	return aaguidMap
}
