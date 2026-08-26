package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

const sessionCookie = "session"

type AuthHandler struct {
	auth         *service.AuthService
	cookieSecure bool
}

func NewAuthHandler(auth *service.AuthService, cookieSecure bool) *AuthHandler {
	return &AuthHandler{auth: auth, cookieSecure: cookieSecure}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Password is required."})
		return
	}
	token, err := h.auth.Login(c.Request.Context(), req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Invalid password."})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	setSessionCookie(c, token, h.cookieSecure)
	c.JSON(http.StatusOK, model.AuthLoginResponse{Authenticated: true})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.auth.Logout(c.Request.Context(), sessionToken(c)); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	clearSessionCookie(c, h.cookieSecure)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Status is public so the SPA can determine its route before authentication.
func (h *AuthHandler) Status(c *gin.Context) {
	resp, err := h.auth.Status(c.Request.Context(), sessionToken(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.CurrentPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Both passwords are required."})
		return
	}
	if err := h.auth.ChangePassword(c.Request.Context(), req.CurrentPassword, req.NewPassword); errors.Is(err, service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Current password is incorrect."})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func sessionToken(c *gin.Context) string {
	token, err := c.Cookie(sessionCookie)
	if err != nil {
		return ""
	}
	return token
}

func setSessionCookie(c *gin.Context, token string, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(sessionCookie, token, int(service.SessionDuration.Seconds()), "/", "", secure, true)
}

func clearSessionCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(sessionCookie, "", -1, "/", "", secure, true)
}

func Authenticated(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, err := auth.Authenticate(c.Request.Context(), sessionToken(c))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
			return
		}
		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, model.ErrorResponse{Error: "Not authenticated."})
			return
		}
		c.Next()
	}
}

func CookieSecureFromEnv(value string) bool {
	if value == "" {
		return true
	}
	secure, err := strconv.ParseBool(value)
	return err != nil || secure
}
