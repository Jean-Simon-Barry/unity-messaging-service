package messaging

type HubMessage struct {
	Sender    uint64
	Receivers []uint64
	Body      []byte
}
