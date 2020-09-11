package user

import (
	"github.com/gin-contrib/sessions"
	"github.com/stretchr/testify/assert"
	"testing"

)

type mockSession struct{}

func (mck *mockSession) GetCurrentUserId() (id uint64, err error) {
	return uint64(10), nil
}

func (mck *mockSession) SetCurrentUser(userId uint64) error {
	return nil
}

func (mck *mockSession) SetSession(session sessions.Session) {}

func TestBooksIndex(t *testing.T) {
	userService := Service{SessionService: &mockSession{}}
	actual :=userService.GetUserId()
	expected := uint64(10)
	assert.Equal(t, expected, actual)
}
