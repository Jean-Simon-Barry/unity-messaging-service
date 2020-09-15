package session

import (
	"errors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type ServiceInterface interface {
	SetCurrentUser(ctx *gin.Context, userId uint64) error
	GetCurrentUserId(ctx *gin.Context) (uint64, error)
}

type sessionService struct{
}

var SessionService ServiceInterface
const userIdKey = "user_id"

func (s *sessionService) SetCurrentUser(c *gin.Context, userId uint64) error {
	session := sessions.Default(c)
	if session == nil {
		return errors.New("no session currently active")
	}
	session.Set(userIdKey, userId)
	return session.Save()
}

func (s *sessionService) GetCurrentUserId(c *gin.Context) (id uint64, err error) {
	session := sessions.Default(c)
	if session == nil {
		return uint64(0), errors.New("no session currently active")
	}
	userId := session.Get(userIdKey)
	if userId == nil {
		return uint64(0), errors.New("no user currently set")
	}
	return userId.(uint64), nil
}

func NewSessionService() {
	SessionService = &sessionService{}
}