package myDB

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() error {
	connStr := "host=localhost user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable port=5432"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS prices (
    product_id TEXT,
    creation_date DATE,  // <--- Проверьте написание!
    product_name TEXT,
    category TEXT,
    price NUMERIC
)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

func InsertPrices(records [][]string) (int, int, float64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO prices (product_id, creation_date, product_name, category, price) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return 0, 0, 0, fmt.Errorf("prepare statement failed: %w", err)
	}
	defer stmt.Close()

	totalItems := 0
	categories := make(map[string]struct{})
	var totalPrice float64

	for _, record := range records {
		if len(record) != 5 {
			continue
		}

		price, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			log.Printf("Invalid price format: %s", record[4])
			continue
		}

		if _, err = stmt.Exec(record[0], record[1], record[2], record[3], price); err != nil {
			log.Printf("Failed to insert record: %v, error: %v", record, err)
			continue
		}

		totalItems++
		categories[record[3]] = struct{}{}
		totalPrice += price
	}

	if err = tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit failed: %w", err)
	}

	return totalItems, len(categories), totalPrice, nil
}

func GetAllPrices() (*sql.Rows, error) {
	return db.Query("SELECT product_id, creation_date, product_name, category, price FROM prices ORDER BY product_id")
}
