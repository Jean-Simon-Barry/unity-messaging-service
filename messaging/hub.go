package messaging

import (
	"unity-messaging-service/redis"
)

type Hub struct {
	// Registered clients mapped by id
	Clients map[uint64]*Client

	// Inbound messages from the clients.
	Relay chan HubMessage

	// Inbound messages from queue.
	QueueMessages chan HubMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

var MessageHub Hub

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.ClientId] = client
			redis.RedisService.CheckUserIn(client.ClientId, RabbitService.GetQueueName())
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ClientId]; ok {
				delete(h.Clients, client.ClientId)
				close(client.Send)
				redis.RedisService.CheckUserOut(client.ClientId)
			}
		case message := <-h.Relay:
			//TODO: check if users are actually online before sending message. If not...put in dead letter queue? save it in another store?
			//TODO: exclude client from sending to himself
			queues := getClientQueues(message.Receivers)
			for queue := range queues {
				RabbitService.PostMessage(queue, message)
			}
		case message := <-h.QueueMessages:
			for _, cid := range message.Receivers {
				if client, ok := h.Clients[cid]; ok {
					client.Send <- message.Body
				}
			}
		}
	}
}

func getClientQueues(clientIds []uint64) map[string]bool {
	if queues, ok := redis.RedisService.GetRabbitQueueNames(clientIds); ok {
		return queues
	}
	return map[string]bool{}
}

func NewHub() *Hub {
	MessageHub = Hub{
		Relay:         make(chan HubMessage),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Clients:       make(map[uint64]*Client),
		QueueMessages: make(chan HubMessage),
	}
	return &MessageHub
}
