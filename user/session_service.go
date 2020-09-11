package user

import (
	"errors"
	"github.com/gin-contrib/sessions"
)

type SessionServiceInterface interface {
	SetCurrentUser(userId uint64) error
	GetCurrentUserId() (uint64, error)
	SetSession(session sessions.Session)
}

type SessionService struct{
	UserSession sessions.Session
}


func (s *SessionService) SetCurrentUser(userId uint64) error {
	if s.UserSession != nil {
		return errors.New("no session currently active")
	}
	s.UserSession.Set("user_id", userId)
	_ = s.UserSession.Save()
	return nil
}

func (s *SessionService) GetCurrentUserId() (id uint64, err error) {
	if s.UserSession == nil {
		return 0, errors.New("no session currently active")
	}
	userId := s.UserSession.Get("user_id")
	if userId == nil {
		return 0, errors.New("no user currently set")
	}
	return userId.(uint64), nil
}

func (s *SessionService) SetSession(currSession sessions.Session) {
	s.UserSession = currSession
}