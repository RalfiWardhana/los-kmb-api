package auth

type Auth struct {
	ClientID   string `json:"client_id" validate:"required"`
	Credential string `json:"credential" validate:"required"`
}

type Authorize struct {
	ClientActive int         `json:"client_active"`
	TokenActive  int         `json:"token_active"`
	Expired      interface{} `json:"expired"`
	AccessToken  string      `json:"access_token"`
}
