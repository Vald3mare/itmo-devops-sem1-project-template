package db

import (
	"database/sql"
)

var database *sql.DB

func InitDB() error {
	// Проверяем подключение
	if err := database.Ping(); err != nil {
		return err
	}

	_, err := database.Exec(`CREATE TABLE IF NOT EXISTS prices (
        id SERIAL PRIMARY KEY,
        product_name TEXT NOT NULL,
        category TEXT NOT NULL,
        price NUMERIC(10,2) NOT NULL,
        creation_date timestamp NOT NULL
    )`)
	return err
}
