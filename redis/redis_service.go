package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"unity-messaging-service/rabbitmq"
)

type DataStore interface {
	GenerateUserId() uint64
	SetRabbitQueue(clientId uint64, queueName string)
	GetRabbitQueueNames(clientIds []uint64) (map[string]bool, bool)
	CheckUserIn(clientId uint64)
	CheckUserOut(clientId uint64)
	GetAllConnectedUsers(caller uint64) []uint64
}

type service struct {
	*redis.Client
}

var ctx = context.Background()
var RedisService DataStore
const activeClientsKey = "activeClients"
const ClientIdKey = "client:id"
const ClientRabbitQueueKey = "client:rabbitQueue"

func buildRabbitQueueKey(clientId uint64) string {
	key := fmt.Sprintf("%s:%s", ClientRabbitQueueKey, strconv.FormatUint(clientId, 10))
	return key
}

func (s *service) GenerateUserId() uint64 {
	incr := s.Client.Incr(ctx, ClientIdKey)
	return uint64(incr.Val())
}


func (s *service) SetRabbitQueue(clientId uint64, queueName string) {
	s.Set(ctx, buildRabbitQueueKey(clientId), queueName, 0)
}

func (s *service) GetRabbitQueueNames(clientIds []uint64) (map[string]bool, bool) {
	queueKeys := make([]string, 0)
	for _, clientId := range clientIds {
		queueKeys = append(queueKeys, buildRabbitQueueKey(clientId))
	}
	r, err := s.MGet(ctx, queueKeys...).Result()
	if err != nil {
		return map[string]bool{}, false
	}
	queueNames := make(map[string]bool)
	for _, queue := range r {
		queueNames[queue.(string)] = true
	}
	return queueNames, true
}

func (s *service) CheckUserIn(clientId uint64) {
	s.SAdd(ctx, activeClientsKey, clientId)
	name := rabbitmq.RabbitService.GetQueueName()
	s.Set(ctx, buildRabbitQueueKey(clientId), name, 0)
}

func (s *service) CheckUserOut(clientId uint64) {
	s.SRem(ctx, activeClientsKey, clientId)
}

func (s *service) GetAllConnectedUsers(caller uint64) []uint64 {
	a := s.SMembers(ctx, activeClientsKey).Val()
	var activeUsers []uint64
	for _, client := range a {
		i, _ := strconv.Atoi(client)
		if caller != uint64(i) {
			activeUsers = append(activeUsers, uint64(i))
		}
	}
	return activeUsers
}

func init() {
	var (
		//TODO: set from env
		host     = "localhost"
		port     = "6379"
		password = ""
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})
	RedisService = &service{client}
}
