package auth

import "time"

// WithTTL sets the time-to-live for tokens.
func WithGetNow(getNow func() time.Time) Option {
	return getNowOption{
		getNow: getNow,
	}
}

type getNowOption struct {
	getNow func() time.Time
}

func (o getNowOption) apply(authenticator *JWTAuthenticator) {
	authenticator.getNow = o.getNow
}
