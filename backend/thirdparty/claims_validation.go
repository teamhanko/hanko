package thirdparty

import (
	"github.com/teamhanko/hanko/backend/v2/utils"
)

// validatePictureURL returns "" if valid, otherwise a stable reason code.
func validatePictureURL(raw string) string {
	return utils.ValidatePictureURL(raw)
}
