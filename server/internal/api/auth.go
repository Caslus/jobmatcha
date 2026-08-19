package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	sessionDuration = 7 * 24 * time.Hour
	tokenBytes      = 32
	bcryptCost      = 12
)

type AuthHandler struct {
	cfgRepo *repository.ConfigRepo
	db      *gorm.DB
}

func NewAuthHandler(cfgRepo *repository.ConfigRepo, db *gorm.DB) *AuthHandler {
	return &AuthHandler{cfgRepo: cfgRepo, db: db}
}

// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	c.BindJSON(&req)

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Password is required."})
		return
	}

	cfg, err := h.cfgRepo.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid password."})
		return
	}

	token, err := h.createSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	setSessionCookie(c, token)
	c.JSON(http.StatusOK, model.AuthTokenResponse{Token: token})
}

// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractToken(c)
	if token != "" {
		h.db.Where("token = ?", token).Delete(&model.Session{})
	}
	clearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GET /api/auth/status
func (h *AuthHandler) Status(c *gin.Context) {
	cfg, err := h.cfgRepo.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	token := extractToken(c)
	authenticated := false
	if token != "" {
		var session model.Session
		authenticated = h.db.Where("token = ? AND expires_at > ?", token, time.Now()).Find(&session).RowsAffected > 0
	}

	resp := model.AuthStatusResponse{
		Authenticated: authenticated,
		SetupComplete: cfg.SetupComplete,
		OIDCEnabled:   cfg.OIDCEnabled,
	}
	if cfg.OIDCEnabled {
		resp.OIDCProviderURL = cfg.OIDCProviderURL
	}

	c.JSON(http.StatusOK, resp)
}

// POST /api/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest
	c.BindJSON(&req)

	if req.CurrentPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Both passwords are required."})
		return
	}

	cfg, err := h.cfgRepo.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Current password is incorrect."})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	if err := h.cfgRepo.UpdateMap(map[string]interface{}{
		"password_hash":  string(hash),
		"setup_complete": true,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *AuthHandler) createSession() (string, error) {
	buf := make([]byte, tokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)

	session := model.Session{
		Token:     token,
		ExpiresAt: time.Now().Add(sessionDuration),
	}
	if err := h.db.Create(&session).Error; err != nil {
		return "", err
	}
	return token, nil
}

func extractToken(c *gin.Context) string {
	// Try cookie first
	token, err := c.Cookie("session")
	if err == nil && token != "" {
		return token
	}
	// Fallback to Authorization: Bearer
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

func setSessionCookie(c *gin.Context, token string) {
	c.SetCookie("session", token, int(sessionDuration.Seconds()), "/", "", false, true)
}

func clearSessionCookie(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
}

// Authenticated returns the auth middleware.
func Authenticated(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Not authenticated."})
			return
		}

		var session model.Session
		result := db.Where("token = ? AND expires_at > ?", token, time.Now()).Find(&session)
		if result.RowsAffected == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Not authenticated."})
			return
		}

		c.Next()
	}
}