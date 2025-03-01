package handlers

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"project-sem/internal/myDB"
)

var db *sql.DB

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := myDB.GetAllPrices()
		if err != nil {
			log.Printf("DB query error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var csvBuffer bytes.Buffer
		writer := csv.NewWriter(&csvBuffer)

		// Write CSV header
		if err := writer.Write([]string{"id", "name", "category", "price", "create_date"}); err != nil {
			log.Printf("CSV header error: %v", err)
			http.Error(w, "CSV generation error", http.StatusInternalServerError)
			return
		}

		// Process rows
		for rows.Next() {
			var id, name, category, date string
			var price float64

			if err := rows.Scan(&id, &name, &category, &price, &date); err != nil {
				log.Printf("Row scan error: %v", err)
				continue
			}

			if err := writer.Write([]string{
				id,
				name,
				category,
				fmt.Sprintf("%.2f", price),
				date,
			}); err != nil {
				log.Printf("CSV write error: %v", err)
				continue
			}
		}
		writer.Flush()

		// Create ZIP in memory
		var zipBuffer bytes.Buffer
		zipWriter := zip.NewWriter(&zipBuffer)
		dataFile, err := zipWriter.Create("data.csv")
		if err != nil {
			log.Printf("ZIP create error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		if _, err := dataFile.Write(csvBuffer.Bytes()); err != nil {
			log.Printf("ZIP write error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		if err := zipWriter.Close(); err != nil {
			log.Printf("ZIP close error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		// Set headers and send response
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", zipBuffer.Len()))
		if _, err := w.Write(zipBuffer.Bytes()); err != nil {
			log.Printf("Response write error: %v", err)
		}
	}
}
