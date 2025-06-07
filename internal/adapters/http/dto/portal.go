package dto

type RequestLogin struct {
	Provider string      `json:"provider"`
	Token    string      `json:"token"`
	Response interface{} `json:"response"`
}
