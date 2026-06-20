package infra

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/billyagasi/ddog-hackathon-octopus/incident-commander/apps/ms-order/internal/config"
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

	return db, nil
}

func migrate(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id UUID PRIMARY KEY,
		customer_id VARCHAR(50) NOT NULL,
		product_id VARCHAR(50) NOT NULL,
		quantity INT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}
	log.Println("[infra] PostgreSQL migration complete")
	return nil
}
