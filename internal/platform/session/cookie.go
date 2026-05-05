// Reads and writes business and admin cookies.

package session

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func (manager *Manager) SetAPISessionCookie(c *gin.Context, claims Claims) error {
	encoded, err := manager.Encode(claims)
	if err != nil {
		return err
	}

	c.SetSameSite(manager.sameSite)
	c.SetCookie(
		manager.cookieName,
		encoded,
		int(manager.maxAge.Seconds()),
		manager.path,
		"",
		manager.secure,
		true,
	)

	return nil
}

func (manager *Manager) ReadAPISessionCookie(c *gin.Context) (Claims, error) {
	value, err := c.Cookie(manager.cookieName)
	if err != nil {
		return Claims{}, ErrSessionCookieMissing
	}

	return manager.Decode(value)
}

func (manager *Manager) ClearAPISessionCookie(c *gin.Context) {
	c.SetSameSite(manager.sameSite)
	c.SetCookie(manager.cookieName, "", -1, manager.path, "", manager.secure, true)
}

func IsMissingSession(err error) bool {
	return errors.Is(err, ErrSessionCookieMissing)
}
