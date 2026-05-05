// Defines session claims, such as `uid`, `role`, and `exp`.

package session

import "time"

type Claims struct {
	UserID      int64  `json:"user_id"`
	Role        string `json:"role"`
	ClassroomID int64  `json:"classroom_id,omitempty"`
	IssuedAt    int64  `json:"iat"`
	ExpiresAt   int64  `json:"exp"`
}

func (claims Claims) Expired(now time.Time) bool {
	return now.UTC().Unix() >= claims.ExpiresAt
}

func (claims Claims) Remaining(now time.Time) time.Duration {
	return time.Unix(claims.ExpiresAt, 0).Sub(now.UTC())
}
