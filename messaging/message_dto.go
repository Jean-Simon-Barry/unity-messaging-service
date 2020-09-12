package messaging

type UserMessage struct {
	Receivers []uint64 `json:"receivers"`
	Message string `json:"message"`
}
