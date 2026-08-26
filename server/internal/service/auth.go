package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	SessionDuration = 7 * 24 * time.Hour
	bcryptCost      = 12
)

type AuthService struct {
	config        *repository.ConfigRepo
	sessions      *repository.SessionRepo
	bootstrapFile string
}

func NewAuthService(repos *repository.Repositories, bootstrapFile string) *AuthService {
	return &AuthService{config: repos.Config, sessions: repos.Session, bootstrapFile: bootstrapFile}
}

func (s *AuthService) EnsureBootstrap(ctx context.Context) error {
	_, err := s.config.Get(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("read config: %w", err)
	}

	password, err := randomSecret(24)
	if err != nil {
		return fmt.Errorf("generate bootstrap password: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash bootstrap password: %w", err)
	}
	cfg := &model.Config{ID: 1, PasswordHash: string(hash), SetupComplete: false, IncludeKeywords: model.StringSlice{}, ExcludeKeywords: model.StringSlice{}, LocationKeywords: model.StringSlice{}, WorkTypes: model.StringSlice{}}
	if err := s.config.Create(ctx, cfg); err != nil {
		return fmt.Errorf("create config: %w", err)
	}
	if err := writeSecretFile(s.bootstrapFile, password+"\n"); err != nil {
		return fmt.Errorf("write bootstrap password file: %w", err)
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, password string) (string, error) {
	cfg, err := s.config.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("read config: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}
	token, err := randomSecret(32)
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	if err := s.sessions.Create(ctx, &model.Session{TokenHash: hashToken(token), ExpiresAt: time.Now().Add(SessionDuration)}); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (s *AuthService) Status(ctx context.Context, token string) (*model.AuthStatusResponse, error) {
	cfg, err := s.config.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	authenticated := false
	if token != "" {
		authenticated, err = s.sessions.Valid(ctx, hashToken(token), time.Now())
		if err != nil {
			return nil, fmt.Errorf("validate session: %w", err)
		}
	}
	resp := &model.AuthStatusResponse{Authenticated: authenticated, SetupComplete: cfg.SetupComplete, OIDCEnabled: cfg.OIDCEnabled}
	if cfg.OIDCEnabled {
		resp.OIDCProviderURL = cfg.OIDCProviderURL
	}
	return resp, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.Delete(ctx, hashToken(token))
}

func (s *AuthService) ChangePassword(ctx context.Context, current, next string) error {
	cfg, err := s.config.Get(ctx)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(current)) != nil {
		return ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(next), bcryptCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.config.UpdateMap(ctx, map[string]interface{}{"password_hash": string(hash)}); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if err := os.Remove(s.bootstrapFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove bootstrap password file: %w", err)
	}
	return nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, nil
	}
	return s.sessions.Valid(ctx, hashToken(token), time.Now())
}

func BootstrapPasswordFile(dbPath string) string {
	if configured := os.Getenv("BOOTSTRAP_PASSWORD_FILE"); configured != "" {
		return configured
	}
	return filepath.Join(filepath.Dir(dbPath), "initial-password")
}

func randomSecret(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writeSecretFile(path, contents string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".initial-password-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0600); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.WriteString(contents); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Link(tempName, path); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}
