package engine

// Response represents a response.
type Response struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// ResponseError represents an error response.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
