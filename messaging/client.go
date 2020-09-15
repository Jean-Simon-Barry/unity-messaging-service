package messaging

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"strconv"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)
const maxMessageSize = 1024

type Client struct {
	Hub *Hub
	Conn *websocket.Conn
	// Buffered channel of outbound messages.
	Send chan []byte
	ClientId uint64
}

func (c *Client) WriteMessages() {
	defer func() {
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					_ = fmt.Errorf("could not close %v", err)
					return
				}
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, err = w.Write(message)
			if err != nil {
				_ = fmt.Errorf("could not write message: %v", err)
			}

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadMessages()  {
	defer func() {
		c.Hub.Unregister <- c
		_ = c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	for {
		var message = &UserMessage{}
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Printf("error: %v", err)
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				break
			}
		} else {
			msgBody := bytes.TrimSpace(bytes.Replace([]byte(message.Message+" [from "+strconv.FormatUint(c.ClientId, 10)+"]"), newline, space, -1))
			hubMessage := HubMessage{Sender: c.ClientId, Receivers: message.Receivers, Body: msgBody}
			c.Hub.ClientMessage <- hubMessage
		}
	}
}