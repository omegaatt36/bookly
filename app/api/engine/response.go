package engine

type response struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
