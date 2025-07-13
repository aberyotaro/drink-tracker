package services

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	DB *sql.DB
}

func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	service := &DatabaseService{DB: db}

	// 環境変数でマイグレーション実行を制御
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := service.runMigrations(); err != nil {
			return nil, err
		}
	}

	return service, nil
}

func (ds *DatabaseService) runMigrations() error {
	migrationFile := "migrations/001_create_tables.sql"

	content, err := os.ReadFile(migrationFile)
	if err != nil {
		return err
	}

	if _, err := ds.DB.Exec(string(content)); err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (ds *DatabaseService) Close() error {
	return ds.DB.Close()
}
