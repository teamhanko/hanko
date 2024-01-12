package dto

type UpdateCredentialDto struct {
	Id   string  `json:"id"`
	Name *string `json:"name,omitempty"`
}
