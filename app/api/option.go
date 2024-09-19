package api

func ptr[T any](v T) *T {
	return &v
}

// Option defines jwt option.
type Option interface {
	apply(*Server)
}

// JWTOption defines a generic jwt option.
type JWTOption struct {
	JWTSalt      string
	JWTSecretKey string
}

func (opt *JWTOption) apply(router *Server) {
	router.jwtSalt = ptr(opt.JWTSalt)
	router.jwtSecret = ptr(opt.JWTSecretKey)
}

// InternalTokenOption defines an option to set internal token.
type InternalTokenOption struct {
	InternalToken string
}

func (opt *InternalTokenOption) apply(router *Server) {
	router.internalToken = ptr(opt.InternalToken)
}

// PortOption defines an option to set port.
type PortOption struct {
	Port int
}

func (opt *PortOption) apply(router *Server) {
	router.port = opt.Port
}
