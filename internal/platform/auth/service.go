// Validates credentials, issues cookies, and handles sign-out.

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"oj-lite/internal/platform/errs"
	"oj-lite/internal/platform/logger"
	"oj-lite/internal/platform/password"
	"oj-lite/internal/platform/user"
)

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errAccountDisabled    = errors.New("account disabled")
	errPasswordMismatch   = errors.New("password mismatch")
)

type loginResult struct {
	User        user.Account
	ClassroomID int64
}

type service struct {
	log   *logger.Logger
	users *user.Store
}

func newService(log *logger.Logger, users *user.Store) *service {
	return &service{
		log:   log,
		users: users,
	}
}

func (service *service) Login(ctx context.Context, username, rawPassword string) (loginResult, error) {
	account, err := service.users.FindByUsername(ctx, strings.TrimSpace(username))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return loginResult{}, errInvalidCredentials
		}

		return loginResult{}, fmt.Errorf("find user by username: %w", err)
	}

	if account.Role != user.RoleTeacher && account.Role != user.RoleStudent {
		return loginResult{}, errInvalidCredentials
	}

	if !account.IsActive() {
		return loginResult{}, errAccountDisabled
	}

	if !password.Verify(account.PasswordHash, rawPassword) {
		return loginResult{}, errInvalidCredentials
	}

	result := loginResult{User: account}
	if account.Role == user.RoleStudent {
		classroomID, err := service.users.FindStudentClassroomID(ctx, account.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return result, nil
			}
			return loginResult{}, fmt.Errorf("find student classroom: %w", err)
		}
		result.ClassroomID = classroomID
	}

	return result, nil
}

func (service *service) GetUser(ctx context.Context, userID int64) (user.Account, error) {
	account, err := service.users.FindByID(ctx, userID)
	if err != nil {
		return user.Account{}, errs.UnauthenticatedIfNoRows(err)
	}

	return account, nil
}

func (service *service) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	account, err := service.users.FindByID(ctx, userID)
	if err != nil {
		return errs.UnauthenticatedIfNoRows(err)
	}

	if !account.IsActive() {
		return errs.Unavailable(errAccountDisabled)
	}

	if !password.Verify(account.PasswordHash, oldPassword) {
		return errs.Unavailable(errPasswordMismatch)
	}

	nextHash, err := password.Hash(newPassword)
	if err != nil {
		if errors.Is(err, password.ErrInvalidLength) {
			return errs.Unavailable(err)
		}
		return err
	}

	if err := service.users.UpdatePasswordHash(ctx, userID, nextHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.Unauthenticated(err)
		}
		return fmt.Errorf("update password hash: %w", err)
	}

	return nil
}
