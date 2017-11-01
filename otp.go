package otp

type OneTimeStore interface {
	NewConn() OneTimeStoreConn
}

type OneTimeStoreConn interface {
	Get(key string) *string
	Exists(key string) bool
	Set(content string, views int, expire int) string
	Close() error
}
