package repository

import (
	"testing"

	"github.com/NexaCard/API/internal/constants"
	"github.com/NexaCard/API/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAdminLoginLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.AdminLoginLog{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestAdminLoginLogRepository_CreateAndList(t *testing.T) {
	db := setupAdminLoginLogTestDB(t)
	repo := NewAdminLoginLogRepository(db)

	entries := []models.AdminLoginLog{
		{AdminID: 1, Username: "admin", EventType: constants.AdminLoginEventLoginPassword, Status: constants.AdminLoginStatusSuccess, ClientIP: "1.1.1.1"},
		{AdminID: 1, Username: "admin", EventType: constants.AdminLoginEventLogin2FAVerify, Status: constants.AdminLoginStatusFailed, FailReason: constants.AdminLoginFailInvalidTOTPCode, ClientIP: "1.1.1.1"},
		{AdminID: 2, Username: "alice", EventType: constants.AdminLoginEventLoginPassword, Status: constants.AdminLoginStatusSuccess, ClientIP: "2.2.2.2"},
	}
	for i := range entries {
		if err := repo.Create(&entries[i]); err != nil {
			t.Fatalf("create: %v", err)
		}
	}

	all, total, err := repo.List(AdminLoginLogListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if total != 3 || len(all) != 3 {
		t.Fatalf("expected 3 records, got total=%d len=%d", total, len(all))
	}

	id := uint(1)
	byAdmin, total, err := repo.List(AdminLoginLogListFilter{AdminID: &id, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list by admin: %v", err)
	}
	if total != 2 || len(byAdmin) != 2 {
		t.Fatalf("expected 2 records for admin 1, got %d/%d", total, len(byAdmin))
	}

	failed, total, err := repo.List(AdminLoginLogListFilter{Status: constants.AdminLoginStatusFailed, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 1 || len(failed) != 1 {
		t.Fatalf("expected 1 failed record, got %d/%d", total, len(failed))
	}
}
