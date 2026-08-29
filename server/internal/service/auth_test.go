package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthServiceBootstrapCreatesOneTimePasswordAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	passwordFile := filepath.Join(t.TempDir(), "secrets", "initial-password")
	service := NewAuthService(repos, passwordFile)

	if err := service.EnsureBootstrap(ctx); err != nil {
		t.Fatalf("EnsureBootstrap: %v", err)
	}
	contents, err := os.ReadFile(passwordFile)
	if err != nil {
		t.Fatalf("read bootstrap password: %v", err)
	}
	password := strings.TrimSpace(string(contents))
	if password == "" {
		t.Fatal("bootstrap password was empty")
	}
	cfg, err := repos.Config.Get(ctx)
	if err != nil {
		t.Fatalf("get bootstrap config: %v", err)
	}
	if cfg.SetupComplete || bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte(password)) != nil {
		t.Fatalf("bootstrap config did not contain the generated password: %#v", cfg)
	}

	if err := service.EnsureBootstrap(ctx); err != nil {
		t.Fatalf("second EnsureBootstrap: %v", err)
	}
	contentsAgain, err := os.ReadFile(passwordFile)
	if err != nil || string(contentsAgain) != string(contents) {
		t.Fatalf("bootstrap password changed: %q, err = %v", contentsAgain, err)
	}
}

func TestAuthServiceBootstrapReportsPasswordFileFailure(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	blockedPath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(blockedPath, []byte("block"), 0600); err != nil {
		t.Fatalf("create blocking file: %v", err)
	}
	service := NewAuthService(testutil.Repositories(db), filepath.Join(blockedPath, "initial-password"))
	if err := service.EnsureBootstrap(ctx); err == nil || !strings.Contains(err.Error(), "write bootstrap password file") {
		t.Fatalf("EnsureBootstrap error = %v, want password file error", err)
	}
}

func TestAuthServiceLoginStatusAndInvalidCredentials(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcryptCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := repos.Config.Create(ctx, &model.Config{ID: 1, PasswordHash: string(hash), SetupComplete: true, OIDCEnabled: true, OIDCProviderURL: "https://issuer.example.test"}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	service := NewAuthService(repos, filepath.Join(t.TempDir(), "initial-password"))

	if _, err := service.Login(ctx, "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login wrong password error = %v, want ErrInvalidCredentials", err)
	}
	status, err := service.Status(ctx, "")
	if err != nil {
		t.Fatalf("Status without token: %v", err)
	}
	if status.Authenticated || !status.SetupComplete || !status.OIDCEnabled || status.OIDCProviderURL != "https://issuer.example.test" {
		t.Fatalf("status without token = %#v", status)
	}
	status, err = service.Status(ctx, "not-a-session")
	if err != nil || status.Authenticated {
		t.Fatalf("Status invalid token = %#v, err = %v", status, err)
	}

	token, err := service.Login(ctx, "correct-password")
	if err != nil || token == "" {
		t.Fatalf("Login valid password = %q, err = %v", token, err)
	}
	status, err = service.Status(ctx, token)
	if err != nil || !status.Authenticated {
		t.Fatalf("Status valid token = %#v, err = %v", status, err)
	}
	if valid, err := service.Authenticate(ctx, token); err != nil || !valid {
		t.Fatalf("Authenticate valid token = %t, err = %v", valid, err)
	}
}

func TestAuthServiceChangePasswordRejectsWrongPasswordAndRemovesBootstrapFile(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	oldHash, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcryptCost)
	if err != nil {
		t.Fatalf("hash old password: %v", err)
	}
	if err := repos.Config.Create(ctx, &model.Config{ID: 1, PasswordHash: string(oldHash)}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	bootstrapFile := filepath.Join(t.TempDir(), "initial-password")
	if err := os.WriteFile(bootstrapFile, []byte("old-password\n"), 0600); err != nil {
		t.Fatalf("create bootstrap password file: %v", err)
	}
	service := NewAuthService(repos, bootstrapFile)

	if err := service.ChangePassword(ctx, "wrong-password", "new-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("ChangePassword wrong password error = %v, want ErrInvalidCredentials", err)
	}
	cfg, err := repos.Config.Get(ctx)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(cfg.PasswordHash), []byte("old-password")) != nil {
		t.Fatalf("wrong password unexpectedly changed config: %#v, err = %v", cfg, err)
	}

	if err := service.ChangePassword(ctx, "old-password", "new-password"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if _, err := os.Stat(bootstrapFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("bootstrap file stat error = %v, want not exist", err)
	}
	if _, err := service.Login(ctx, "old-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login old password error = %v, want ErrInvalidCredentials", err)
	}
	if _, err := service.Login(ctx, "new-password"); err != nil {
		t.Fatalf("Login new password: %v", err)
	}
}

func TestAuthServiceStatusRejectsExpiredSessionAndMissingConfig(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	service := NewAuthService(repos, filepath.Join(t.TempDir(), "initial-password"))
	if _, err := service.Status(ctx, "token"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Status without config error = %v, want ErrNotFound", err)
	}
	if _, err := service.Login(ctx, "password"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("Login without config error = %v, want ErrNotFound", err)
	}
	if err := repos.Config.Create(ctx, &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	if err := repos.Session.Create(ctx, &model.Session{TokenHash: hashToken("expired"), ExpiresAt: time.Now().Add(-time.Minute)}); err != nil {
		t.Fatalf("create expired session: %v", err)
	}
	status, err := service.Status(ctx, "expired")
	if err != nil || status.Authenticated {
		t.Fatalf("Status expired token = %#v, err = %v", status, err)
	}
}
