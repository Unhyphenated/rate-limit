package models

type RateLimitResult struct {
	Allowed bool
	Limit int64
	Remaining int64
	ResetAt int64
	RetryAfter int64
	FailOpen bool
}