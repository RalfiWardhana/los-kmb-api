package dto

type AuthModel struct {
	Lob        string `json:"lob" validate:"required,lob"`
	Channel    string `json:"channel" validate:"required,channel"`
	ClientID   string `json:"client_id" validate:"required"`
	Credential string `json:"credential" validate:"required"`
}

type AuthJoinTable struct {
	ClientActive int         `json:"client_active"`
	TokenActive  int         `json:"token_active"`
	Expired      interface{} `json:"expired"`
	AccessToken  string      `json:"access_token"`
}

type AuthTools struct {
	ClientID   string `json:"client_id" validate:"required"`
	Credential string `json:"credential" validate:"required"`
}
