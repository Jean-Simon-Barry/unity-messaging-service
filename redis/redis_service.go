package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"sync"
)

type DataStore interface {
	GenerateUserId() uint64
	GetRabbitQueueNames(clientIds []uint64) (map[string]bool, bool)
	CheckUserIn(clientId uint64, queueName string)
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
var lock sync.Mutex

func buildRabbitQueueKey(clientId uint64) string {
	key := fmt.Sprintf("%s:%s", ClientRabbitQueueKey, strconv.FormatUint(clientId, 10))
	return key
}

func (s *service) GenerateUserId() uint64 {
	lock.Lock()
	defer lock.Unlock()
	incr := s.Client.Incr(ctx, ClientIdKey)
	log.Printf("generated new client id %d", incr.Val())
	return uint64(incr.Val())
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
		if queue != nil {
			queueNames[queue.(string)] = true
		}
	}
	return queueNames, true
}

func (s *service) CheckUserIn(clientId uint64, queueName string) {
	log.Printf("logging user %d into queue %s\n", clientId, queueName)
	s.SAdd(ctx, activeClientsKey, clientId)
	s.Set(ctx, buildRabbitQueueKey(clientId), queueName, 0)
}

func (s *service) CheckUserOut(clientId uint64) {
	log.Printf("logging user %d out\n", clientId)
	s.Del(ctx, buildRabbitQueueKey(clientId))
	s.SRem(ctx, activeClientsKey, clientId)
}

func (s *service) GetAllConnectedUsers(caller uint64) []uint64 {
	log.Println("fetching connected users")
	a := s.SMembers(ctx, activeClientsKey).Val()
	var activeUsers []uint64
	for _, client := range a {
		i, _ := strconv.Atoi(client)
		if caller != uint64(i) {
			activeUsers = append(activeUsers, uint64(i))
		}
	}
	log.Printf("connected users are %v\n", activeUsers)
	return activeUsers
}

func NewRedisService() {
	var (
		//TODO: set from env
		host     = "unity-msg-svc-redis-master"
		port     = "6379"
		password = ""
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("%s: %s", "could not connect to redis", err)
	}
	RedisService = &service{client}
}
