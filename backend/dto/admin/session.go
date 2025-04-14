package admin

type CreateSessionTokenDto struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	UserAgent string `json:"user_agent"`
	IpAddress string `json:"ip_address" validate:"omitempty,ip"`
}

type CreateSessionTokenResponse struct {
	SessionToken string `json:"session_token"`
}

type ListSessionsRequestDto struct {
	UserID string `param:"user_id" validate:"required,uuid"`
}

type DeleteSessionRequestDto struct {
	ListSessionsRequestDto
	SessionID string `param:"session_id" validate:"required,uuid4"`
}
