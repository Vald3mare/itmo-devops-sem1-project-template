package handlers

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
)

var db *sql.DB

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Явно указываем названия полей и формат даты
		rows, err := db.Query(`
		SELECT 
			product_id AS id,
			TO_CHAR(creation_date, 'YYYY-MM-DD') AS creation_date,
			product_name AS name,
			category,
			price
		FROM prices
		ORDER BY id
	`)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Создаем CSV с нужными заголовками
		csvFile, err := os.CreateTemp("", "data-*.csv")
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(csvFile.Name())
		defer csvFile.Close()

		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		// Заголовки согласно тестовым данным
		headers := []string{"id", "name", "category", "price", "create_date"}
		if err := writer.Write(headers); err != nil {
			http.Error(w, "CSV error", http.StatusInternalServerError)
			return
		}

		// Заполняем данные
		for rows.Next() {
			var id, date, name, category string
			var price float64

			if err := rows.Scan(&id, &date, &name, &category, &price); err != nil {
				continue
			}

			record := []string{
				id,
				name,
				category,
				fmt.Sprintf("%.2f", price),
				date, // Здесь date будет в колонке create_date
			}

			if err := writer.Write(record); err != nil {
				continue
			}
		}

		// Создаем ZIP архив
		zipFile, err := os.CreateTemp("", "data-*.zip")
		if err != nil {
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(zipFile.Name())
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		dataFile, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		csvContent, _ := os.ReadFile(csvFile.Name())
		if _, err := dataFile.Write(csvContent); err != nil {
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		http.ServeFile(w, r, zipFile.Name())
	}
}
