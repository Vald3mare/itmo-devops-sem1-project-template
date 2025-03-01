package handlers

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var db *sql.DB

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
        SELECT 
            product_id,
            TO_CHAR(creation_date, 'YYYY-MM-DD'),
            product_name,
            category,
            price
        FROM prices
        ORDER BY product_id
    `)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Создаем временный CSV файл
		csvFile, err := os.CreateTemp("", "data-*.csv")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(csvFile.Name())
		defer csvFile.Close()

		// Создаем CSV writer
		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		// Записываем заголовок
		if err := writer.Write([]string{"id", "creation_date", "product_name", "category", "price"}); err != nil {
			http.Error(w, "CSV generation failed", http.StatusInternalServerError)
			return
		}

		// Обрабатываем каждую запись
		for rows.Next() {
			var id, date, name, category string
			var price float64

			if err := rows.Scan(&id, &date, &name, &category, &price); err != nil {
				continue
			}

			record := []string{
				id,
				date,
				strings.TrimSpace(name),
				strings.TrimSpace(category),
				fmt.Sprintf("%.2f", price),
			}

			if err := writer.Write(record); err != nil {
				continue
			}
		}

		// Проверяем ошибки итерации
		if err := rows.Err(); err != nil {
			http.Error(w, "Data processing error", http.StatusInternalServerError)
			return
		}

		// Создаем ZIP архив
		zipFile, err := os.CreateTemp("", "data-*.zip")
		if err != nil {
			http.Error(w, "Archive creation failed", http.StatusInternalServerError)
			return
		}
		defer os.Remove(zipFile.Name())
		defer zipFile.Close()

		// Создаем ZIP writer
		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// Добавляем CSV файл в архив
		dataFile, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		// Копируем данные в архив
		csvContent, _ := os.ReadFile(csvFile.Name())
		if _, err := dataFile.Write(csvContent); err != nil {
			http.Error(w, "Archive write error", http.StatusInternalServerError)
			return
		}

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")

		// Отправляем файл
		http.ServeFile(w, r, zipFile.Name())
	}
}
