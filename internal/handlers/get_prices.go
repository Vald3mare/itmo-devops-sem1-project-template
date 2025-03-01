package handlers

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
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
    `)
		if err != nil {
			log.Printf("DB query error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Создаем временный CSV файл
		csvFile, err := os.CreateTemp("", "data-*.csv")
		if err != nil {
			log.Printf("Temp file error: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(csvFile.Name())
		defer csvFile.Close()

		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		// Пишем заголовок
		if err := writer.Write([]string{"id", "creation_date", "product_name", "category", "price"}); err != nil {
			log.Printf("Header write error: %v", err)
			http.Error(w, "CSV error", http.StatusInternalServerError)
			return
		}

		// Обрабатываем строки
		for rows.Next() {
			var id, date, name, category string
			var price float64

			if err := rows.Scan(&id, &date, &name, &category, &price); err != nil {
				log.Printf("Row scan error: %v", err)
				continue
			}

			record := []string{id, date, name, category, fmt.Sprintf("%.2f", price)}
			if err := writer.Write(record); err != nil {
				log.Printf("Record write error: %v", err)
				continue
			}
		}

		// Проверяем ошибки после итерации
		if err := rows.Err(); err != nil {
			log.Printf("Rows iteration error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Создаем ZIP архив
		zipFile, err := os.CreateTemp("", "data-*.zip")
		if err != nil {
			log.Printf("Zip create error: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		defer os.Remove(zipFile.Name())
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// Добавляем CSV в архив
		dataFile, err := zipWriter.Create("data.csv")
		if err != nil {
			log.Printf("Zip entry create error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		// Копируем данные CSV в архив
		csvContent, err := os.ReadFile(csvFile.Name())
		if err != nil {
			log.Printf("CSV read error: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		if _, err := dataFile.Write(csvContent); err != nil {
			log.Printf("Zip write error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		// Важно закрыть writer перед отправкой файла
		zipWriter.Close()

		// Устанавливаем заголовки
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")

		// Отправляем файл
		http.ServeFile(w, r, zipFile.Name())
	}
}
