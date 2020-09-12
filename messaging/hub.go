package messaging

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
		case client := <-h.Unregister:
			if _, ok := h.Clients[client.ClientId]; ok {
				delete(h.Clients, client.ClientId)
				close(client.Send)
			}
		case message := <-h.Relay:
			for _, cid := range message.receivers {
				if client, ok := h.Clients[cid]; ok {
					client.Send <- message.msg
				}
			}
		}
	}
}

var MessageHub Hub

func NewHub() *Hub {
	return &Hub{
		Relay:      make(chan HubMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[uint64]*Client),
	}
}
