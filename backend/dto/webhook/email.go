package webhook

type EmailSend struct {
	Subject          string `json:"subject"`        // subject
	BodyPlain        string `json:"body_plain"`     // used for string templates
	Body             string `json:"body,omitempty"` // used for HTML templates
	ToEmailAddress   string `json:"to_email_address"`
	DeliveredByHanko bool   `json:"delivered_by_hanko"`
	AcceptLanguage   string `json:"accept_language"` // Deprecated. Accept-Language header from HTTP request
	Language         string `json:"language"`        // X-Language header from HTTP request
	Type             string `json:"type"`            // type of the email

	Data interface{} `json:"data"`
}

type PasscodeData struct {
	ServiceName string `json:"service_name"`
	OtpCode     string `json:"otp_code"`
	TTL         int    `json:"ttl"`
	ValidUntil  int64  `json:"valid_until"` // UnixTimestamp
}
