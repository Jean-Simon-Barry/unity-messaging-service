package user

import (
	"unity-messaging-service/redis"
)

type ServiceInterface interface {
	GetUserId() uint64
}

type Service struct {
	SessionService SessionServiceInterface
}

func (s *Service) GetUserId() uint64 {
	id, err := s.SessionService.GetCurrentUserId()
	if err != nil {
		generatedId := redis.RedisService.GenerateUserId()
		s.SessionService.SetCurrentUser(generatedId)
		return generatedId
	} else {
		return id
	}
}
