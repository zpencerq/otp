package otp

type OneTimeStore interface {
	Get(key string) *string
	Exists(key string) bool
	Set(content string, views int, expire int) string
}
