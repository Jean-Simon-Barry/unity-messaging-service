package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unity-messaging-service/messaging"
	"unity-messaging-service/mocks"
	"unity-messaging-service/redis"
	"unity-messaging-service/session"
)

func TestListUsersEndpoint(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	router := setupRouter()

	MockRedis := mocks.NewMockDataStore(controller)
	MockSession := mocks.NewMockSessionServiceInterface(controller)

	redis.RedisService = MockRedis
	session.SessionService = MockSession

	MockRedis.EXPECT().GetAllConnectedUsers(uint64(0)).Return([]uint64{1234})
	MockSession.EXPECT().GetCurrentUserId(gomock.Any()).Return(uint64(0), nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"users\":[1234]}", w.Body.String())
}

func TestIdentityEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := gomock.NewController(t)
	defer controller.Finish()
	MockSession := mocks.NewMockSessionServiceInterface(controller)
	session.SessionService = MockSession

	router := setupRouter()

	w := httptest.NewRecorder()
	MockSession.EXPECT().GetCurrentUserId(gomock.Any()).Return(uint64(1234), nil)

	req, _ := http.NewRequest("GET", "/identity", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{\"user_id\":1234}", w.Body.String())
}

func TestIdentityEndpointNoUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()
	controller := gomock.NewController(t)
	defer controller.Finish()
	MockSession := mocks.NewMockSessionServiceInterface(controller)
	session.SessionService = MockSession

	w := httptest.NewRecorder()
	expected := errors.New("no session currently active")
	MockSession.EXPECT().GetCurrentUserId(gomock.Any()).Return(uint64(0), expected)

	req, _ := http.NewRequest("GET", "/identity", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "{\"error\":\"" + expected.Error() + "\"}", w.Body.String())
}

func TestRelayHandler(t *testing.T) {
	const userId = uint64(1234)
	gin.SetMode(gin.TestMode)
	controller := gomock.NewController(t)
	defer controller.Finish()
	MockSession := mocks.NewMockSessionServiceInterface(controller)
	MockRedis := mocks.NewMockDataStore(controller)
	MockRabbit := mocks.NewMockRabbitInterface(controller)
	redis.RedisService = MockRedis
	session.SessionService = MockSession
	messaging.RabbitService = MockRabbit
	MockRedis.EXPECT().CheckUserIn(userId, "queue")
	MockRabbit.EXPECT().GetQueueName().Return("queue")
	MockSession.EXPECT().GetCurrentUserId(gomock.Any()).Return(userId, nil)
	messaging.NewHub()

	router := setupRouter()
	ts := httptest.NewServer(router)
	defer ts.Close()
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + "/relay"
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	_ = ws.WriteJSON(&messaging.UserMessage{Receivers: []uint64{userId}, Message: "hello!"})
	defer ws.Close()

	assert.Containsf(t, messaging.MessageHub.Clients, userId, "contains client")
	//assert that channel was used to send message?
}
