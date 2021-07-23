package auth

import (
	"github.com/gin-contrib/sessions"
)

func (m *Module) clearSession(s sessions.Session) {
	s.Clear()

	// newOptions := router.SessionOptions(m.config)
	// newOptions.MaxAge = -1 // instruct browser to delete cookie immediately
	// s.Options(newOptions)

	if err := s.Save(); err != nil {
		panic(err)
	}
}
