// Provides data access for user records, including lookups by username and id.

package user

import (
	"context"
	"database/sql"

	"oj-lite/internal/platform/logger"
)

type Store struct {
	db  *sql.DB
	log *logger.Logger
}

func NewStore(database *sql.DB, log *logger.Logger) *Store {
	return &Store{
		db:  database,
		log: log,
	}
}

func (store *Store) FindByUsername(ctx context.Context, username string) (Account, error) {
	return store.scanAccount(store.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, status, created_at
		FROM user_account
		WHERE username = ?
		LIMIT 1
	`, username))
}

func (store *Store) FindByID(ctx context.Context, userID int64) (Account, error) {
	return store.scanAccount(store.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, role, status, created_at
		FROM user_account
		WHERE id = ?
		LIMIT 1
	`, userID))
}

func (store *Store) FindStudentClassroomID(ctx context.Context, studentID int64) (int64, error) {
	var classroomID int64
	err := store.db.QueryRowContext(ctx, `
		SELECT classroom_id
		FROM enrollment
		WHERE student_id = ?
		LIMIT 1
	`, studentID).Scan(&classroomID)
	if err != nil {
		return 0, err
	}

	return classroomID, nil
}

func (store *Store) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	result, err := store.db.ExecContext(ctx, `
		UPDATE user_account
		SET password_hash = ?
		WHERE id = ?
	`, passwordHash, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (store *Store) scanAccount(scanner interface{ Scan(dest ...any) error }) (Account, error) {
	var account Account
	err := scanner.Scan(
		&account.ID,
		&account.Username,
		&account.PasswordHash,
		&account.Role,
		&account.Status,
		&account.CreatedAt,
	)
	if err != nil {
		return Account{}, err
	}

	return account, nil
}
