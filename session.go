package libserver

type Session interface {
	Id() string
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
	Has(key string) bool
	IsExpired() bool
	Update()
	Clear()
}
