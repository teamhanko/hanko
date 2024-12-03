package admin

type ListWebauthnCredentialsRequestDto struct {
	UserID string `param:"user_id" validate:"required,uuid"`
}

type GetWebauthnCredentialRequestDto struct {
	ListWebauthnCredentialsRequestDto
	WebauthnCredentialID string `param:"credential_id" validate:"required"`
}
