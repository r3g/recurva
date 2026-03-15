package service

import (
	"testing"

	"github.com/r3g/recurva/internal/domain"
	"github.com/r3g/recurva/internal/store/memory"
)

func TestTagService_AddAndList(t *testing.T) {
	svc := NewTagService(*memory.New())

	tag, err := svc.AddTag(ctx(), "gre")
	if err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if tag.Name != "gre" {
		t.Errorf("Name = %q, want %q", tag.Name, "gre")
	}
	if tag.ID == "" {
		t.Error("expected non-empty ID")
	}

	if _, err := svc.AddTag(ctx(), "sat"); err != nil {
		t.Fatalf("AddTag sat: %v", err)
	}

	tags, err := svc.ListTags(ctx())
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("len(tags) = %d, want 2", len(tags))
	}
}

func TestTagService_AddDuplicate(t *testing.T) {
	svc := NewTagService(*memory.New())

	if _, err := svc.AddTag(ctx(), "gre"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	_, err := svc.AddTag(ctx(), "gre")
	if err != domain.ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestTagService_AddEmpty(t *testing.T) {
	svc := NewTagService(*memory.New())

	_, err := svc.AddTag(ctx(), "")
	if err != domain.ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTagService_Rename(t *testing.T) {
	svc := NewTagService(*memory.New())

	if _, err := svc.AddTag(ctx(), "biz"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if err := svc.RenameTag(ctx(), "biz", "business"); err != nil {
		t.Fatalf("RenameTag: %v", err)
	}

	// Old name should not exist
	tags, _ := svc.ListTags(ctx())
	for _, tag := range tags {
		if tag.Name == "biz" {
			t.Error("old tag name 'biz' still exists")
		}
		if tag.Name == "business" {
			return // success
		}
	}
	t.Error("renamed tag 'business' not found")
}

func TestTagService_RenameNotFound(t *testing.T) {
	svc := NewTagService(*memory.New())

	err := svc.RenameTag(ctx(), "nonexistent", "new")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestTagService_Delete(t *testing.T) {
	svc := NewTagService(*memory.New())

	if _, err := svc.AddTag(ctx(), "gre"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}
	if err := svc.DeleteTag(ctx(), "gre"); err != nil {
		t.Fatalf("DeleteTag: %v", err)
	}
	tags, _ := svc.ListTags(ctx())
	if len(tags) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(tags))
	}
}

func TestTagService_DeleteNotFound(t *testing.T) {
	svc := NewTagService(*memory.New())

	err := svc.DeleteTag(ctx(), "nonexistent")
	if err != domain.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

