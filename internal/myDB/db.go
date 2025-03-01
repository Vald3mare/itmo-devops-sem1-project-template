package myDB

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB инициализирует подключение к базе данных
func InitDB() error {
	connStr := "host=localhost user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable port=5432"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS prices (
			id TEXT,
			name TEXT,
			category TEXT,
			price NUMERIC,
			create_date DATE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// InsertPrices вставляет данные в базу
func InsertPrices(records [][]string) (int, int, float64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("transaction error: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO prices(id, name, category, price, create_date)
		VALUES($1, $2, $3, $4, $5)
	`)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("prepare error: %w", err)
	}
	defer stmt.Close()

	totalItems := 0
	categories := make(map[string]struct{})
	var totalPrice float64

	for _, record := range records {
		if len(record) != 5 {
			continue
		}

		price, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			continue
		}

		_, err = stmt.Exec(
			record[0], // id
			record[1], // name
			record[2], // category
			price,
			record[4], // create_date
		)
		if err != nil {
			log.Printf("Insert error: %v", err)
			continue
		}

		totalItems++
		categories[record[2]] = struct{}{}
		totalPrice += price
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit error: %w", err)
	}

	return totalItems, len(categories), totalPrice, nil
}

// GetAllPrices возвращает все записи
func GetAllPrices() (*sql.Rows, error) {
	return db.Query(`
		SELECT 
			id,
			name,
			category,
			price,
			TO_CHAR(create_date, 'YYYY-MM-DD')
		FROM prices
		ORDER BY id
	`)
}
