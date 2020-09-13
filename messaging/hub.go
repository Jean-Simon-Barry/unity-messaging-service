package messaging

import (
	"fmt"
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
			queues := getClientQueues(message.receivers)
			for _, queue := range queues {
				fmt.Println(queue)
			}
			for _, cid := range message.receivers {
				if client, ok := h.Clients[cid]; ok {
					client.Send <- message.msg
				}
			}
		}
	}
}

func (h *Hub) GetConnectedClients(caller uint64) []uint64 {
	keys := make([]uint64, 0, len(h.Clients))
	for k, _ := range h.Clients {
		if k != caller {
			keys = append(keys, k)
		}
	}
	return keys
}

func getClientQueues(clientIds []uint64) []string {
	queueNames := make(map[string]bool)
	for _, cid := range clientIds {
		if name, ok := redis.RedisService.GetRabbitQueueName(cid); ok {
			queueNames[name] = true
		} else {
			//TODO: figure out what to do when a user is not logged in. Should do anything? Drop in dead letter queue?
		}
	}
	distinctQueues := make([]string, len(queueNames))
	i := 0
	for k := range queueNames {
		distinctQueues[i] = k
		i++
	}
	return distinctQueues
}

func NewHub() *Hub {
	return &Hub{
		Relay:      make(chan HubMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[uint64]*Client),
	}
}
