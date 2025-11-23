package libserver

type Session interface {
	Get(key string) any
	Set(key string, value any)
	Delete(key string)
	Has(key string) bool
	Clear()
}
