package messaging

type HubMessage struct {
	Sender    uint64   `json:"sender"`
	Receivers []uint64 `json:"receivers"`
	Body      []byte   `json:"body"`
}
