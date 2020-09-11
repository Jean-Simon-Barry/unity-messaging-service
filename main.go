package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"unity-messaging-service/user"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.StaticFile("/favicon.ico", "./resources/favicon.ico")
	var sessionService user.SessionServiceInterface
	var userService user.ServiceInterface

	sessionService = &user.SessionService{}
	userService = &user.Service{SessionService: sessionService}
	r.Use(sessions.Sessions("user_id", cookie.NewStore([]byte("secret"))))

	r.GET("/connect", func(c *gin.Context) {
		sessionService.SetSession(sessions.Default(c))
		userId := userService.GetUserId()
		c.JSON(200, gin.H{"user_id": userId})
	})
	return r
}

func main() {
	r := setupRouter()
	_ = r.Run()
}
