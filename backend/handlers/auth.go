package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
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

var store = session.New()

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

type AuthStatusRequest struct {
	SessionID string `cookie:"session_id"`
}

type AuthStatusResponse struct {
	Body AuthStatusBody
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

	c := utils.GetFiberCtx(ctx)

	// Get or create session
	s, _ := store.Get(c)

	// If this is a new session
	if s.Fresh() {
		// Save session data
		s.Set("username", input.Body.Username)
		err := s.Save()
		if err != nil {
			log.Println(err)
		}
	}

	return &LoginResponse{
		Body: AuthSuccessBody{
			Success: true,
			Message: "Login successful",
		},
	}, nil
}

// LogoutHandler handles user logout (Huma handler)
func LogoutHandler(ctx context.Context, input *LogoutRequest) (*struct{}, error) {
	c := utils.GetFiberCtx(ctx)

	s, _ := store.Get(c)

	s.Destroy()

	return nil, nil
}

func isAuthenticated(c *fiber.Ctx) bool {
	// .Get creates a new session if one does not exist
	s, err := store.Get(c)

	if err != nil || s.Fresh() {
		// delete the session if one was just created automatically by .Get
		if s.Fresh() {
			s.Destroy()
		}
		return false
	}

	return true
}

// AuthStatusHandler returns the current authentication status (Huma handler)
func AuthStatusHandler(ctx context.Context, input *AuthStatusRequest) (*AuthStatusResponse, error) {
	c := utils.GetFiberCtx(ctx)

	return &AuthStatusResponse{
		Body: AuthStatusBody{
			Authenticated: isAuthenticated(c),
		},
	}, nil
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !isAuthenticated(c) {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		return c.Next()
	}
}
