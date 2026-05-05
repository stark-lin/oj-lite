// Defines user entity structures stored in the database.

package user

const (
	RoleTeacher = "teacher"
	RoleStudent = "student"

	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type Account struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

func (account Account) IsActive() bool {
	return account.Status == StatusActive
}
