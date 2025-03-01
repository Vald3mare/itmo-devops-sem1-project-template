package handlers

import (
	"archive/zip"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"

	"project-sem/internal/myDB"
)

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := myDB.GetAllPrices()
		if err != nil {
			http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		csvFile, err := os.CreateTemp("", "data-*.csv")
		if err != nil {
			http.Error(w, "Error creating csv: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(csvFile.Name())
		defer csvFile.Close()

		writer := csv.NewWriter(csvFile)
		defer writer.Flush()

		for rows.Next() {
			var id, date, name, category string
			var price float64
			if err := rows.Scan(&id, &date, &name, &category, &price); err != nil {
				http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := writer.Write([]string{id, date, name, category, fmt.Sprintf("%.2f", price)}); err != nil {
				http.Error(w, "Error writing csv: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error after scanning: "+err.Error(), http.StatusInternalServerError)
			return
		}

		zipFile, err := os.CreateTemp("", "data-*.zip")
		if err != nil {
			http.Error(w, "Error creating zip: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(zipFile.Name())
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		dataFile, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "Error creating zip entry: "+err.Error(), http.StatusInternalServerError)
			return
		}

		csvContent, err := os.ReadFile(csvFile.Name())
		if err != nil {
			http.Error(w, "Error reading csv: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err = dataFile.Write(csvContent); err != nil {
			http.Error(w, "Error writing to zip: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		http.ServeFile(w, r, zipFile.Name())
	}
}
