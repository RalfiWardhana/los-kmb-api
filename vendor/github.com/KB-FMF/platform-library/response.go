package platform

import "strconv"

// Response is default platform response
type Response struct {
	Messages string                 `json:"messages"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Errors   interface{}            `json:"errors"`
	Code     string                 `json:"code"`
}

func (r Response) CodeInt() int {
	n, _ := strconv.Atoi(r.Code)
	return n
}
