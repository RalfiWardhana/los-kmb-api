package usecase

type ErrorValidation2Wilen struct {
	Message   string `json:"message" example:"Parameter prospect_id Required field"`
	Parameter string `json:"parameter" example:"prospect_id"`
}

type ErrorValidationResponse2Wilen struct {
	Code       string                `json:"code" example:"LOS-WLN-800"`
	Data       interface{}           `json:"data,omitempty"`
	Error      ErrorValidation2Wilen `json:"errors,omitempty"`
	Message    string                `json:"message" example:"validasi data kamu ada yang tidak sesuai, silakan periksa kembali"`
	MetaData   interface{}           `json:"metadata,omitempty"`
	ServerTime string                `json:"server_time" example:"2024-11-14 14:16:43.619763 +0700 WIB"`
	XRequestID string                `json:"x-request-id,omitempty" example:"80980b86-6800-49c6-8c6a-d2049c280be0"`
}

type ErrorResponse2Wilen struct {
	Code       string      `json:"code" example:"LOS-WLN-999"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"errors,omitempty" example:"internal server error"`
	Message    string      `json:"message" example:"Terjadi kesalahan pada sistem, silahkan coba lagi"`
	MetaData   interface{} `json:"metadata,omitempty"`
	ServerTime string      `json:"server_time" example:"2024-11-14 14:16:43.619763 +0700 WIB"`
	XRequestID string      `json:"x-request-id,omitempty" example:"80980b86-6800-49c6-8c6a-d2049c280be0"`
}

type SuccessResponse2Wilen struct {
	Code       string      `json:"code" example:"LOS-WLN-001"`
	Data       interface{} `json:"data"`
	Error      interface{} `json:"errors,omitempty"`
	Message    string      `json:"message" example:"ok"`
	MetaData   interface{} `json:"metadata,omitempty"`
	ServerTime string      `json:"server_time" example:"2024-11-14 14:16:43.619763 +0700 WIB"`
	XRequestID string      `json:"x-request-id,omitempty" example:"80980b86-6800-49c6-8c6a-d2049c280be0"`
}
