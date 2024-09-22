package web

// Option defines jwt option.
type Option interface {
	apply(*Server)
}

// PortOption defines an option to set port.
type PortOption struct {
	Port int
}

func (opt *PortOption) apply(router *Server) {
	router.port = opt.Port
}

// ServerURLOption defines an option to set server url.
type ServerURLOption struct {
	ServerURL string
}

func (opt *ServerURLOption) apply(router *Server) {
	router.serverURL = opt.ServerURL
}
