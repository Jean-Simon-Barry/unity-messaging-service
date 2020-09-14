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
	r.GET("/list", listHandler)
	r.GET("/health-check", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})
	return r
}
/*
	when users land on home page of the chat, an id is generated for them and set as cookie in their session
	then we render the homepage back to them.
 */
func homeHandler(c *gin.Context) {
	cid := redis.RedisService.GenerateUserId()
	_ = session.SessionService.SetCurrentUser(c, cid)
	c.HTML(200, "index.html", nil)
}
/*
	handles the web socket chat requests by registering the user into the hub (using their generated user id). Then we launch 2 goroutines
	to read and write from/to their socket connection.
 */
func wsHandler(hub *messaging.Hub, c *gin.Context) {
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("failed to set websocket upgrade", err)
		return
	}

	cid, err := session.SessionService.GetCurrentUserId(c)
	if err != nil {
		fmt.Println("failed to set websocket upgrade", err)
		_ = wsConn.Close()
	}
	userClient := &messaging.Client{Hub:hub, Conn:wsConn, Send:make(chan []byte, 256), ClientId:cid}
	userClient.Hub.Register <- userClient

	go userClient.ReadMessages()
	go userClient.WriteMessages()
	_ = wsConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("You are client with id: %d", cid)))
}
/*
	lists all the users connected to the hub, except the current user calling the endpoint.
 */
func listHandler(c *gin.Context) {
	cid, _ := session.SessionService.GetCurrentUserId(c)
	users := redis.RedisService.GetAllConnectedUsers(cid)
	if users == nil {
		users = []uint64{}
	}
	c.JSON(200, gin.H{"users": users})
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
