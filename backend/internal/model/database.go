package model

import (
	"log/slog"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes the SQLite database and runs auto-migrations.
func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&Admin{},
		&Server{},
		&Image{},
		&Application{},
		&SystemConfig{},
	); err != nil {
		return nil, err
	}
	if err := repairLegacyImagesSchema(db); err != nil {
		return nil, err
	}

	slog.Info("Database initialized", "path", dbPath)
	return db, nil
}

type pragmaTableInfo struct {
	Name string `gorm:"column:name"`
	Type string `gorm:"column:type"`
}

type pragmaForeignKey struct {
	Table string `gorm:"column:table"`
	From  string `gorm:"column:from"`
	To    string `gorm:"column:to"`
}

func repairLegacyImagesSchema(db *gorm.DB) error {
	if !db.Migrator().HasTable("images") {
		return nil
	}

	var columns []pragmaTableInfo
	if err := db.Raw("PRAGMA table_info(images)").Scan(&columns).Error; err != nil {
		return err
	}

	var imageIDType string
	for _, column := range columns {
		if column.Name == "image_id" {
			imageIDType = strings.ToUpper(column.Type)
			break
		}
	}

	var foreignKeys []pragmaForeignKey
	if err := db.Raw("PRAGMA foreign_key_list(images)").Scan(&foreignKeys).Error; err != nil {
		return err
	}

	needsRepair := imageIDType != "" && imageIDType != "TEXT"
	for _, fk := range foreignKeys {
		if fk.From == "image_id" && fk.Table == "applications" {
			needsRepair = true
			break
		}
	}
	if !needsRepair {
		return nil
	}

	slog.Warn("Repairing legacy images table schema")

	if err := db.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return err
	}
	defer db.Exec("PRAGMA foreign_keys = ON")

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DROP TABLE IF EXISTS images_new").Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			CREATE TABLE images_new (
				id integer PRIMARY KEY AUTOINCREMENT,
				server_id integer NOT NULL,
				image_id text NOT NULL,
				name text NOT NULL,
				image_address text NOT NULL,
				created_at datetime,
				CONSTRAINT fk_images_server FOREIGN KEY (server_id) REFERENCES servers(id)
			)
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec(`
			INSERT INTO images_new (id, server_id, image_id, name, image_address, created_at)
			SELECT id, server_id, CAST(image_id AS TEXT), name, image_address, created_at
			FROM images
		`).Error; err != nil {
			return err
		}
		if err := tx.Exec("DROP TABLE images").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE images_new RENAME TO images").Error; err != nil {
			return err
		}
		if err := tx.Exec("CREATE INDEX idx_images_server_id ON images(server_id)").Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
