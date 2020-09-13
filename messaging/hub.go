package messaging

import (
	"encoding/json"
	"unity-messaging-service/rabbitmq"
	"unity-messaging-service/redis"
)

type Hub struct {
	// Registered clients mapped by id
	Clients map[uint64]*Client

	// Inbound messages from the clients.
	Relay chan HubMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ClientId] = client
			redis.RedisService.CheckUserIn(client.ClientId)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ClientId]; ok {
				delete(h.Clients, client.ClientId)
				close(client.Send)
				redis.RedisService.CheckUserOut(client.ClientId)
			}
		case message := <-h.Relay:
			queues := getClientQueues(message.Receivers)
			for queue := range queues {
				jsonMessage, _ := json.Marshal(message)
				rabbitmq.RabbitService.PostMessage(queue, jsonMessage)
			}
			for _, cid := range message.Receivers {
				if client, ok := h.Clients[cid]; ok {
					client.Send <- message.Body
				}
			}
		}
	}
}

func getClientQueues(clientIds []uint64) map[string]bool {
	queueNames := make(map[string]bool)
	for _, cid := range clientIds {
		if name, ok := redis.RedisService.GetRabbitQueueName(cid); ok {
			queueNames[name] = true
		} else {
			//TODO: figure out what to do when a user is not logged in. Should do anything? Drop in dead letter queue?
		}
	}
	return queueNames
}

func NewHub() *Hub {
	return &Hub{
		Relay:      make(chan HubMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[uint64]*Client),
	}
}
