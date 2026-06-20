package infra

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-inventory/internal/config"
)

func NewPostgresDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("[infra] PostgreSQL connected")

	if err := migrate(db); err != nil {
		return nil, err
	}

	if err := seed(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS inventory (
		product_id VARCHAR(50) PRIMARY KEY,
		stock INT NOT NULL DEFAULT 0,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	log.Println("[infra] PostgreSQL migration complete")
	return nil
}

func seed(db *sql.DB) error {
	products := []struct {
		id    string
		stock int
	}{
		{"SKU-001", 100},
		{"SKU-002", 200},
		{"SKU-003", 50},
		{"SKU-004", 300},
		{"SKU-005", 75},
	}

	for _, p := range products {
		_, err := db.Exec(
			`INSERT INTO inventory (product_id, stock, updated_at) VALUES ($1, $2, NOW())
			 ON CONFLICT (product_id) DO NOTHING`,
			p.id, p.stock,
		)
		if err != nil {
			return err
		}
	}
	log.Println("[infra] inventory seeded")
	return nil
}
