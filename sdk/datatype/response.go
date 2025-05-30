package datatype

import "encoding/json"

type Response struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type LoginUserResponse struct {
	Token string `json:"token"`
}
