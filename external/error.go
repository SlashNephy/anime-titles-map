package external

import "errors"

var (
	ErrRateLimited = errors.New("rate limited")
	ErrServerError = errors.New("server error")
)
