package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

func (m *Module) clearSession(s sessions.Session) {
	for _, key := range sessionKeys {
		s.Delete(key)
	}

	newOptions := router.SessionOptions(m.config)
	newOptions.MaxAge = -1 // instruct browser to delete cookie immediately
	s.Options(newOptions)

	if err := s.Save(); err != nil {
		panic(err)
	}
}
