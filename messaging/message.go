package messaging

type HubMessage struct {
	sender uint64
	receivers []uint64
	msg []byte
}
