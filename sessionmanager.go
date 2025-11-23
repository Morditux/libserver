package libserver

type SessionManager interface {
	CreateSession() Session
	GetSession(id string) Session
	DeleteSession(id string)
}
