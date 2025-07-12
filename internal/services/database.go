package services

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseService struct {
	DB *sql.DB
}

func NewDatabaseService(dbPath string) (*DatabaseService, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	service := &DatabaseService{DB: db}

	if err := service.runMigrations(); err != nil {
		return nil, err
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
