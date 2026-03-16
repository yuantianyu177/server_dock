package model

import (
	"path/filepath"
	"strings"
	"testing"
)

type tableInfoRow struct {
	Name string `gorm:"column:name"`
	Type string `gorm:"column:type"`
}

type foreignKeyRow struct {
	Table string `gorm:"column:table"`
	From  string `gorm:"column:from"`
}

func TestInitDB_CreatesCorrectImagesSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "serverdock.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	var columns []tableInfoRow
	if err := db.Raw("PRAGMA table_info(images)").Scan(&columns).Error; err != nil {
		t.Fatalf("PRAGMA table_info(images) failed: %v", err)
	}

	var imageIDType string
	for _, column := range columns {
		if column.Name == "image_id" {
			imageIDType = strings.ToUpper(column.Type)
			break
		}
	}
	if imageIDType != "TEXT" {
		t.Fatalf("Expected images.image_id to be TEXT, got %q", imageIDType)
	}

	var foreignKeys []foreignKeyRow
	if err := db.Raw("PRAGMA foreign_key_list(images)").Scan(&foreignKeys).Error; err != nil {
		t.Fatalf("PRAGMA foreign_key_list(images) failed: %v", err)
	}

	for _, fk := range foreignKeys {
		if fk.From == "image_id" {
			t.Fatalf("images.image_id should not be a foreign key, got reference to %s", fk.Table)
		}
	}
}
