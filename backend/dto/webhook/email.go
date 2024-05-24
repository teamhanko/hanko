package webhook

import "github.com/teamhanko/hanko/backend/persistence/models"

type EmailSend struct {
	Subject          string    `json:"subject"`        // subject
	BodyPlain        string    `json:"body_plain"`     // used for string templates
	Body             string    `json:"body,omitempty"` // used for html templates
	ToEmailAddress   string    `json:"to_email_address"`
	DeliveredByHanko bool      `json:"delivered_by_hanko"`
	AcceptLanguage   string    `json:"accept_language"` // accept_language header from http request
	Type             EmailType `json:"type"`            // type of the email, currently only "passcode", but other could be added later

	Data interface{} `json:"data"`
}

type PasscodeData struct {
	ServiceName string `json:"service_name"`
	OtpCode     string `json:"otp_code"`
	TTL         int    `json:"ttl"`
	ValidUntil  int64  `json:"valid_until"` // UnixTimestamp
}

type PasslinkData struct {
	ServiceName  string                    `json:"service_name"`
	Token        string                    `json:"token"`
	URL          string                    `json:"url"`
	TTL          int                       `json:"ttl"`
	ValidUntil   int64                     `json:"valid_until"` // UnixTimestamp
	RedirectPath string                    `json:"redirect_path"`
	RetryLimit   int                       `json:"retry_limit"`
	Strictness   models.PasslinkStrictness `json:"strictness"`
}

type EmailType string

var (
	EmailTypePasscode EmailType = "passcode"
	EmailTypePasslink EmailType = "passlink"
)
