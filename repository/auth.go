package repository

import (
	"fmt"
	"invoice-generator-backend/config"
	"strings"
	"time"
)

func CreateUser(name, email, passwordHash string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	var exists bool
	err := config.DB.QueryRow(
		`SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE lower(trim(email)) = lower(trim($1))
		)`,
		normalizedEmail,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrDuplicateEmail
	}

	query := `
		INSERT INTO users (
			name, email, password, created_by, updated_by
		)
		VALUES ($1, $2, $3, $2, $2)
	`

	_, err = config.DB.Exec(query, name, normalizedEmail, passwordHash)
	return err
}

func GetUserPasswordHashByEmail(email string) (string, error) {
	query := `
		SELECT password
		FROM users
		WHERE lower(email) = lower($1)
	`

	var passwordHash string
	err := config.DB.QueryRow(query, email).Scan(&passwordHash)
	if err != nil {
		return "", err
	}

	return passwordHash, nil
}

func GetUserNameByEmail(email string) (string, error) {
	query := `
		SELECT name
		FROM users
		WHERE lower(email) = lower($1)
	`

	var name string
	err := config.DB.QueryRow(query, email).Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil
}

func GetUserProfileByEmail(email string) (string, string, error) {
	query := `
		SELECT name, email
		FROM users
		WHERE lower(email) = lower($1)
	`

	var name string
	var actualEmail string
	err := config.DB.QueryRow(query, email).Scan(&name, &actualEmail)
	if err != nil {
		return "", "", err
	}

	return name, actualEmail, nil
}

func CreateLoginSession(email, token string, loggedInAt time.Time, expiresIn int, expiresAt time.Time) error {
	loggedInEpoch := loggedInAt.Unix()
	expiresInEpoch := expiresAt.Unix()

	query := `
		INSERT INTO login_sessions (email, token, logged_in, expires_in, expires_at, logged_in_epoch, expires_in_epoch)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := config.DB.Exec(query, email, token, loggedInAt, expiresIn, expiresAt, loggedInEpoch, expiresInEpoch)
	return err
}

func CreatePasswordResetToken(email, token string, expiresAt time.Time) error {
	// Ensure a user can request multiple resets by replacing any existing token rows for that email.
	deleteQuery := `
		DELETE FROM password_reset_tokens
		WHERE lower(email) = lower($1)
	`
	if _, err := config.DB.Exec(deleteQuery, email); err != nil {
		return fmt.Errorf("delete existing reset tokens: %w", err)
	}

	insertWithAuditQuery := `
		INSERT INTO password_reset_tokens (
			email, token, expires_at, created_by, updated_by
		)
		VALUES ($1, $2, $3, $1, $1)
	`

	if _, err := config.DB.Exec(insertWithAuditQuery, email, token, expiresAt); err == nil {
		return nil
	} else {
		// Fallback for databases where audit columns are not present.
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "created_by") || strings.Contains(errText, "updated_by") {
			insertBaseQuery := `
				INSERT INTO password_reset_tokens (email, token, expires_at)
				VALUES ($1, $2, $3)
			`
			if _, fallbackErr := config.DB.Exec(insertBaseQuery, email, token, expiresAt); fallbackErr != nil {
				return fmt.Errorf("insert reset token (fallback) failed: %w", fallbackErr)
			}
			return nil
		}

		return fmt.Errorf("insert reset token failed: %w", err)
	}
}

func VerifyPasswordResetToken(token string) (string, error) {
	query := `
		SELECT email
		FROM password_reset_tokens
		WHERE token = $1 AND expires_at > NOW()
	`

	var email string
	err := config.DB.QueryRow(query, token).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

func DeletePasswordResetToken(token string) error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE token = $1
	`

	_, err := config.DB.Exec(query, token)
	return err
}

func UpdateUserPassword(email, passwordHash string) error {
	query := `
		UPDATE users
		SET password = $1, updated_by = $2
		WHERE lower(email) = lower($2)
	`

	_, err := config.DB.Exec(query, passwordHash, email)
	return err
}
