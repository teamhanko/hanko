package webhook

type EmailSend struct {
	Subject          string    `json:"subject"`        // subject
	BodyPlain        string    `json:"body_plain"`     // used for string templates
	Body             string    `json:"body,omitempty"` // used for HTML templates
	ToEmailAddress   string    `json:"to_email_address"`
	DeliveredByHanko bool      `json:"delivered_by_hanko"`
	AcceptLanguage   string    `json:"accept_language"` // Deprecated. Accept-Language header from HTTP request
	Language         string    `json:"language"`        // X-Language header from HTTP request
	Type             EmailType `json:"type"`            // type of the email, currently only "passcode", but other could be added later

	Data interface{} `json:"data"`
}

type PasscodeData struct {
	ServiceName string `json:"service_name"`
	OtpCode     string `json:"otp_code"`
	TTL         int    `json:"ttl"`
	ValidUntil  int64  `json:"valid_until"` // UnixTimestamp
}

type EmailType string

func EmailTypeFromStashedTemplateName(stashedTemplateName string) EmailType {
	switch stashedTemplateName {
	case "login":
		return EmailTypePasscode
	case "email_login_attempted":
		return EmailTypeLoginAttempt
	case "email_verification":
		return EmailTypeEmailVerification
	case "email_registration_attempted":
		return EmailTypeRegistrationAttempt
	case "recovery":
		return EmailTypeRecovery
	default:
		return EmailTypeUnknown
	}
}

var (
	EmailTypePasscode            EmailType = "passcode"
	EmailTypeLoginAttempt        EmailType = "login_attempt"
	EmailTypeRegistrationAttempt EmailType = "registration_attempt"
	EmailTypeEmailVerification   EmailType = "email_verification"
	EmailTypeRecovery            EmailType = "recovery"
	EmailTypeUnknown             EmailType = "unknown"
)
