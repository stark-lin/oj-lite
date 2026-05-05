// Provides helper functions for retrieving the current user from context.

package auth

import "github.com/gin-gonic/gin"

const contextKeyCurrentUser = "current_user"

type CurrentUser struct {
	ID          int64  `json:"id,omitempty"`
	Username    string `json:"username,omitempty"`
	Role        string `json:"role"`
	ClassroomID int64  `json:"classroom_id,omitempty"`
}

func SetCurrentUser(c *gin.Context, user CurrentUser) {
	c.Set(contextKeyCurrentUser, user)
}

func GetCurrentUser(c *gin.Context) (CurrentUser, bool) {
	value, ok := c.Get(contextKeyCurrentUser)
	if !ok {
		return CurrentUser{}, false
	}

	user, ok := value.(CurrentUser)
	if !ok {
		return CurrentUser{}, false
	}

	return user, true
}
