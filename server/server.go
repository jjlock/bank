package server

import (
	"github.com/jjlock/bank/server/db"
)

type server struct {
	db      db.Store
	session sessionStore
}

func New() *server {
	return &server{
		db:      make(db.Memcache),
		session: newCookieStore(),
	}
}

func (s *server) LoadDB(file string) error {
	return s.db.Load(file)
}

func (s *server) SaveDB(file string) error {
	return s.db.Save(file)
}
