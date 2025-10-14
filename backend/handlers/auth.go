package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3/v2"
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
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Username string `json:"username"`
}

type AuthStatusBody struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
}

func setupSessionStore() *session.Store {
	var isDevelopment = utils.IsDevelopment()

	storage := sqlite3.New(sqlite3.Config{
		Database: utils.EnsureDbPath(),
		Table:    "auth_sessions",
		Reset:    false,
	})

	var store = session.New(session.Config{
		CookieSecure:   !isDevelopment,
		CookieHTTPOnly: true,
		Storage:        storage,
	})

	return store
}

var store = setupSessionStore()

type LoginRequest struct {
	Body struct {
		Username string `json:"username" validate:"required" doc:"Username"`
		Password string `json:"password" validate:"required" doc:"Password"`
	}
}

type LoginResponse struct {
	Status    int
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
		return nil, huma.Error401Unauthorized("Invalid username or password")
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
			Success:  true,
			Message:  "Login successful",
			Username: input.Body.Username,
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

func isAuthenticated(c *fiber.Ctx) (*session.Session, bool) {
	// .Get creates a new session if one does not exist
	s, err := store.Get(c)

	if err != nil || s.Fresh() {
		// delete the session if one was just created automatically by .Get
		if s.Fresh() {
			s.Destroy()
		}
		return s, false
	}

	return s, true
}

// AuthStatusHandler returns the current authentication status (Huma handler)
func AuthStatusHandler(ctx context.Context, input *AuthStatusRequest) (*AuthStatusResponse, error) {
	c := utils.GetFiberCtx(ctx)
	session, authenticated := isAuthenticated(c)

	username := ""

	if authenticated {
		username = session.Get("username").(string)
	}

	return &AuthStatusResponse{
		Body: AuthStatusBody{
			Authenticated: authenticated,
			Username:      username,
		},
	}, nil
}

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, authenticated := isAuthenticated(c)

		if !authenticated {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		return c.Next()
	}
}
