// Defines login, password-change, and user-info response structures.

package auth

import "oj-lite/internal/platform/user"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type userDTO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func newUserDTO(account user.Account) userDTO {
	return userDTO{
		ID:        account.ID,
		Username:  account.Username,
		Role:      account.Role,
		Status:    account.Status,
		CreatedAt: account.CreatedAt,
	}
}
