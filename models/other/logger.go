package other

type CustomLog struct {
	ProspectID string      `json:"prospect_id"`
	Info       interface{} `json:"info,omitempty"`
	Error      interface{} `json:"error,omitempty"`
}

type ResultLog struct {
	Request  interface{} `json:"request,omitempty"`
	Response interface{} `json:"response,omitempty"`
}
