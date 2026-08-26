package repository

import (
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUpsertTailoredSerializesDocumentOnUpdate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	if err := db.AutoMigrate(&model.Resume{}, &model.TailoredResume{}); err != nil {
		t.Fatalf("migrating database: %v", err)
	}

	repo := NewResumeRepo(db)
	resume := &model.Resume{Filename: "resume.txt", Content: "original"}
	if err := repo.Create(resume); err != nil {
		t.Fatalf("creating resume: %v", err)
	}
	if err := repo.UpsertTailored(&model.TailoredResume{
		ResumeID: resume.ID,
		RoleID:   1,
		Document: model.ResumeDocument{Content: "first version"},
	}); err != nil {
		t.Fatalf("creating tailored resume: %v", err)
	}
	if err := repo.UpsertTailored(&model.TailoredResume{
		ResumeID: resume.ID,
		RoleID:   1,
		Document: model.ResumeDocument{Content: "updated version"},
	}); err != nil {
		t.Fatalf("updating tailored resume: %v", err)
	}

	tailored, err := repo.GetTailored(resume.ID, 1)
	if err != nil {
		t.Fatalf("loading tailored resume: %v", err)
	}
	if tailored == nil || tailored.Document.Content != "updated version" {
		t.Fatalf("saved document = %#v, want updated version", tailored)
	}

	var count int64
	if err := db.Model(&model.TailoredResume{}).Count(&count).Error; err != nil {
		t.Fatalf("counting tailored resumes: %v", err)
	}
	if count != 1 {
		t.Fatalf("tailored resume count = %d, want 1", count)
	}
}

func TestUpdateDocumentSerializesStructuredResume(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	if err := db.AutoMigrate(&model.Resume{}); err != nil {
		t.Fatalf("migrating database: %v", err)
	}

	repo := NewResumeRepo(db)
	resume := &model.Resume{Filename: "resume.txt", Content: "original"}
	if err := repo.Create(resume); err != nil {
		t.Fatalf("creating resume: %v", err)
	}

	want := model.ResumeDocument{
		Header:   model.ResumeHeader{Name: "Ada Lovelace", Contact: []string{"ada@example.com"}},
		Sections: []model.ResumeSection{{Heading: "Experience", Kind: "experience"}},
	}
	if err := repo.UpdateDocument(resume.ID, want); err != nil {
		t.Fatalf("updating structured resume: %v", err)
	}

	updated, err := repo.GetLatest()
	if err != nil {
		t.Fatalf("loading updated resume: %v", err)
	}
	if updated == nil || updated.Document.Header.Name != want.Header.Name || len(updated.Document.Sections) != 1 {
		t.Fatalf("saved document = %#v, want %#v", updated, want)
	}
}
