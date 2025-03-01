package myDB

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

// InitDB инициализирует подключение к базе данных
func InitDB() error {
	connStr := "user=validator password=val1dat0r dbname=project-sem-1 sslmode=disable port=5432"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	// Создание таблицы если не существует
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS prices (
		product_id TEXT,
		creation_date DATE,
		product_name TEXT,
		category TEXT,
		price NUMERIC
	)`)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

// InsertPrices вставляет данные из CSV в базу данных
func InsertPrices(records [][]string) (int, int, float64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, 0, 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO prices (product_id, creation_date, product_name, category, price) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return 0, 0, 0, err
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
			continue
		}

		_, err = stmt.Exec(record[0], record[1], record[2], record[3], price)
		if err != nil {
			continue
		}

		totalItems++
		categories[record[3]] = struct{}{}
		totalPrice += price
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, err
	}

	return totalItems, len(categories), totalPrice, nil
}

// GetAllPrices возвращает все записи из базы данных
func GetAllPrices() (*sql.Rows, error) {
	return db.Query("SELECT product_id, creation_date, product_name, category, price FROM prices")
}
