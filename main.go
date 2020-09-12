package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"unity-messaging-service/messaging"
	"unity-messaging-service/redis"
	"unity-messaging-service/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	hub := messaging.NewHub()
	go hub.Run()
	r := setupRouter(hub)
	r.LoadHTMLFiles("index.html")
	_ = r.Run()
}

func setupRouter(hub *messaging.Hub) *gin.Engine {
	r := gin.Default()
	r.StaticFile("/favicon.ico", "./resources/favicon.ico")
	r.Use(sessions.Sessions("user_id", cookie.NewStore([]byte("secret"))))
	r.GET("/", homeHandler)
	r.GET("/relay", func(context *gin.Context) {
		wsHandler(hub, context)
	})
	r.GET("/identity", identityHandler)
	return r
}

func homeHandler(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func wsHandler(hub *messaging.Hub, c *gin.Context) {
	cid := redis.RedisService.GenerateUserId()
	_ = session.SessionService.SetCurrentUser(c, cid)
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("failed to set websocket upgrade", err)
		return
	}

	userClient := &messaging.Client{Hub:hub, Conn:wsConn, Send:make(chan []byte, 256), ClientId:cid}
	userClient.Hub.Register <- userClient

	go userClient.ReadMessages()
	go userClient.WriteMessages()
	_ = wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Welcome! You are client with id: %d", cid)))
}

func identityHandler(c *gin.Context) {
	cid, err := session.SessionService.GetCurrentUserId(c)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{"user_id": cid})
}
