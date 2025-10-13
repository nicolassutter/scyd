package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolassutter/scyd/utils"
	"golang.org/x/crypto/bcrypt"
)

type Session struct {
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Response body structs
type AuthSuccessBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type AuthStatusBody struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
}

var (
	sessions      = make(map[string]*Session)
	sessionMutex  = &sync.RWMutex{}
	sessionExpiry = 24 * time.Hour // Sessions expire after 24 hours
)

type LoginRequest struct {
	Body struct {
		Username string `json:"username" validate:"required" doc:"Username"`
		Password string `json:"password" validate:"required" doc:"Password"`
	}
}

type LoginResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      AuthSuccessBody
}

type LogoutRequest struct {
	SessionID string `cookie:"session_id"`
}

type LogoutResponse struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

type AuthStatusRequest struct {
	SessionID string `cookie:"session_id"`
}

type AuthStatusResponse struct {
	Body AuthStatusBody
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// func generatePasswordHash(password string) (string, error) {
// 	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(hash), nil
// }

// creates a new session and returns the session ID
func createSession(username string) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	sessions[sessionID] = &Session{
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionExpiry),
	}

	return sessionID, nil
}

func getSession(sessionID string) (*Session, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()

	session, exists := sessions[sessionID]
	// Check if session exists and is not expired
	if !exists || time.Now().After(session.ExpiresAt) {
		return nil, false
	}

	return session, true
}

// deleteSession removes a session
func deleteSession(sessionID string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessions, sessionID)
}

func cleanupExpiredSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	now := time.Now()
	for id, session := range sessions {
		if now.After(session.ExpiresAt) {
			delete(sessions, id)
		}
	}
}

// LoginHandler handles user login (Huma handler)
func LoginHandler(ctx context.Context, input *LoginRequest) (*LoginResponse, error) {
	// Check if user exists and password is correct
	user, exists := utils.UserConfig.Users[input.Body.Username]
	if !exists || !verifyPassword(input.Body.Password, user.PasswordHash) {
		return &LoginResponse{
			Body: AuthSuccessBody{
				Success: false,
				Message: "Invalid username or password",
			},
		}, nil
	}

	// Create session
	sessionID, err := createSession(input.Body.Username)
	if err != nil {
		return &LoginResponse{
			Body: AuthSuccessBody{
				Success: false,
				Message: "Failed to create session",
			},
		}, nil
	}

	// Set session cookie
	cookie := http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Expires:  time.Now().Add(sessionExpiry),
		HttpOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}

	return &LoginResponse{
		SetCookie: cookie,
		Body: AuthSuccessBody{
			Success: true,
			Message: "Login successful",
		},
	}, nil
}

// LogoutHandler handles user logout (Huma handler)
func LogoutHandler(ctx context.Context, input *LogoutRequest) (*LogoutResponse, error) {
	// Get session ID from cookie
	if input.SessionID != "" {
		deleteSession(input.SessionID)
	}

	// Clear cookie by setting it to expire in the past
	clearCookie := http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Unix(0, 0), // Expired
		HttpOnly: true,
		Secure:   !utils.IsDevelopment(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1, // Delete immediately
	}

	return &LogoutResponse{
		SetCookie: clearCookie,
	}, nil
}

// AuthStatusHandler returns the current authentication status (Huma handler)
func AuthStatusHandler(ctx context.Context, input *AuthStatusRequest) (*AuthStatusResponse, error) {
	// Check if user is authenticated
	if input.SessionID != "" {
		if session, exists := getSession(input.SessionID); exists {
			return &AuthStatusResponse{
				Body: AuthStatusBody{
					Authenticated: true,
					Username:      session.Username,
				},
			}, nil
		}
	}

	return &AuthStatusResponse{
		Body: AuthStatusBody{
			Authenticated: false,
		},
	}, nil
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies("session_id")
		if sessionID == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		session, exists := getSession(sessionID)
		if !exists {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Store user info in context for use in handlers
		c.Locals("username", session.Username)
		return c.Next()
	}
}

// StartSessionCleanup starts a goroutine to periodically clean up expired sessions
func StartSessionCleanup() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
		defer ticker.Stop()

		for range ticker.C {
			cleanupExpiredSessions()
		}
	}()
}
