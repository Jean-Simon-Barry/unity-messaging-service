package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"unity-messaging-service/redis"
	"unity-messaging-service/session"
)

func main() {
	r := setupRouter()
	_ = r.Run()
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.StaticFile("/favicon.ico", "./resources/favicon.ico")
	r.Use(sessions.Sessions("user_id", cookie.NewStore([]byte("secret"))))
	r.GET("/", homeHandler)
	r.GET("/identity", identityHandler)
	return r
}

func homeHandler(c *gin.Context) {
	cid := redis.RedisService.GenerateUserId()
	_ = session.SessionService.SetCurrentUser(c, cid)
	c.JSON(200, gin.H{"user_id": cid})
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
