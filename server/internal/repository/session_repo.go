package repository

import (
	"context"
	"errors"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type SessionRepo struct{ db *gorm.DB }

func NewSessionRepo(db *gorm.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *SessionRepo) Valid(ctx context.Context, tokenHash string, now time.Time) (bool, error) {
	var session model.Session
	err := r.db.WithContext(ctx).Where("token_hash = ? AND expires_at > ?", tokenHash, now).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return err == nil, err
}

func (r *SessionRepo) Delete(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&model.Session{}).Error
}
