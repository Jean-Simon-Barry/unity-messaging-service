package messaging

import (
	mapset "github.com/deckarep/golang-set"
	"log"
	"unity-messaging-service/redis"
)

type Hub struct {
	// Registered clients mapped by id
	Clients map[uint64]*Client

	// Inbound messages from the clients.
	ClientMessage chan HubMessage

	// Inbound messages from queue.
	QueueMessages chan HubMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

var MessageHub Hub

func toInterface(slice []uint64) []interface{} {
	b := make([]interface{}, len(slice))
	for i := range slice {
		b[i] = slice[i]
	}
	return b
}

func getOnlineTargetClients(message HubMessage) []uint64 {
	targetReceivers := mapset.NewSetFromSlice(toInterface(message.Receivers))
	targetReceivers.Remove(message.Sender)
	connectedUsers := mapset.NewSetFromSlice(toInterface(redis.RedisService.GetAllConnectedUsers(message.Sender)))
	onLineClients := connectedUsers.Union(targetReceivers).ToSlice()
	ids := make([]uint64, len(onLineClients))
	for _, c := range onLineClients {
		ids = append(ids, c.(uint64))
	}
	return ids
}

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
		case message := <-h.ClientMessage:
			//TODO: What to do if client is offline ? Put in dead letter queue? save it in another store?
			log.Printf("received new relay request from %d to clients %v", message.Sender, message.Receivers)
			clients := getOnlineTargetClients(message)
			targetQueues := getClientQueues(clients)
			for queue := range targetQueues {
				RabbitService.PostMessage(queue, message)
			}
		case message := <-h.QueueMessages:
			log.Printf("received new message from queue request from %d to clients %v", message.Sender, message.Receivers)
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

func NewHub() {
	MessageHub = Hub{
		ClientMessage: make(chan HubMessage),
		Register:      make(chan *Client),
		Unregister:    make(chan *Client),
		Clients:       make(map[uint64]*Client),
		QueueMessages: make(chan HubMessage),
	}
	go MessageHub.Run()
}
