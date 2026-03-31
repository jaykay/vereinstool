package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/jaykay/vereinstool/backend/config"
	"github.com/jaykay/vereinstool/backend/db/generated"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrSessionExpired     = errors.New("session expired or invalid")
	ErrInvalidResetToken  = errors.New("invalid or expired reset token")
)

const (
	bcryptCost     = 12
	sessionTTL     = 30 * 24 * time.Hour // 30 days
	resetTokenTTL  = 1 * time.Hour
	resetTokenSize = 32 // bytes
)

type Auth struct {
	queries *generated.Queries
	cfg     *config.Config
}

func NewAuth(queries *generated.Queries, cfg *config.Config) *Auth {
	return &Auth{queries: queries, cfg: cfg}
}

// HashPassword hashes a plaintext password with bcrypt.
func (a *Auth) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func (a *Auth) CheckPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// Login verifies credentials and creates a session. Returns session ID.
func (a *Auth) Login(ctx context.Context, email, password string) (string, *generated.User, error) {
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, fmt.Errorf("querying user: %w", err)
	}

	if user.IsActive == 0 {
		return "", nil, ErrUserInactive
	}

	if err := a.CheckPassword(user.PasswordHash, password); err != nil {
		return "", nil, ErrInvalidCredentials
	}

	sessionID := uuid.New().String()
	_, err = a.queries.CreateSession(ctx, generated.CreateSessionParams{
		ID:        sessionID,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(sessionTTL),
	})
	if err != nil {
		return "", nil, fmt.Errorf("creating session: %w", err)
	}

	return sessionID, &user, nil
}

// Logout deletes a session.
func (a *Auth) Logout(ctx context.Context, sessionID string) error {
	return a.queries.DeleteSession(ctx, sessionID)
}

// ValidateSession checks if a session is valid and returns the associated user.
func (a *Auth) ValidateSession(ctx context.Context, sessionID string) (*generated.User, error) {
	session, err := a.queries.GetSession(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionExpired
		}
		return nil, fmt.Errorf("querying session: %w", err)
	}

	user, err := a.queries.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("querying user for session: %w", err)
	}

	if user.IsActive == 0 {
		return nil, ErrUserInactive
	}

	return &user, nil
}

// CreatePasswordResetToken generates a reset token for the given email.
// Returns the token string. Does NOT return an error if user is not found (to prevent enumeration).
func (a *Auth) CreatePasswordResetToken(ctx context.Context, email string) (string, error) {
	user, err := a.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // silent: don't reveal if user exists
		}
		return "", fmt.Errorf("querying user: %w", err)
	}

	tokenBytes := make([]byte, resetTokenSize)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	err = a.queries.CreatePasswordResetToken(ctx, generated.CreatePasswordResetTokenParams{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(resetTokenTTL),
	})
	if err != nil {
		return "", fmt.Errorf("storing reset token: %w", err)
	}

	return token, nil
}

// ResetPassword validates the token and sets a new password.
func (a *Auth) ResetPassword(ctx context.Context, token, newPassword string) error {
	resetToken, err := a.queries.GetPasswordResetToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrInvalidResetToken
		}
		return fmt.Errorf("querying reset token: %w", err)
	}

	hash, err := a.HashPassword(newPassword)
	if err != nil {
		return err
	}

	err = a.queries.UpdateUserPassword(ctx, generated.UpdateUserPasswordParams{
		PasswordHash: hash,
		ID:           resetToken.UserID,
	})
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	_ = a.queries.MarkPasswordResetTokenUsed(ctx, token)
	_ = a.queries.DeleteUserSessions(ctx, resetToken.UserID)

	return nil
}

// SeedAdmin creates the initial admin user if it doesn't exist.
func (a *Auth) SeedAdmin(ctx context.Context, email, password string) error {
	_, err := a.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return nil // already exists
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("checking admin existence: %w", err)
	}

	hash, err := a.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = a.queries.CreateUser(ctx, generated.CreateUserParams{
		Email:        email,
		Name:         "Admin",
		PasswordHash: hash,
		Role:         "admin",
	})
	if err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}

	log.Printf("Seeded admin user: %s", email)
	return nil
}

// StartSessionCleanup periodically removes expired sessions.
func (a *Auth) StartSessionCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.queries.DeleteExpiredSessions(ctx); err != nil {
				log.Printf("Session cleanup error: %v", err)
			}
			if err := a.queries.DeleteExpiredPasswordResetTokens(ctx); err != nil {
				log.Printf("Reset token cleanup error: %v", err)
			}
		}
	}
}
