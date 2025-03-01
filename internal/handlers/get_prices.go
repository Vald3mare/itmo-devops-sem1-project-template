package handlers

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
)

var db *sql.DB

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
            SELECT 
                product_id,
                product_name,
                category,
                price,
                TO_CHAR(creation_date, 'YYYY-MM-DD') 
            FROM prices
            ORDER BY product_id
        `)
		if err != nil {
			log.Printf("Database error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var csvBuffer bytes.Buffer
		writer := csv.NewWriter(&csvBuffer)

		// Заголовок согласно требованиям
		writer.Write([]string{"id", "name", "category", "price", "create_date"})

		for rows.Next() {
			var id, name, category, date string
			var price float64

			if err := rows.Scan(&id, &name, &category, &price, &date); err != nil {
				log.Printf("Row scan error: %v", err)
				continue
			}

			writer.Write([]string{
				id,
				name,
				category,
				fmt.Sprintf("%.2f", price),
				date,
			})
		}
		writer.Flush()

		var zipBuffer bytes.Buffer
		zipWriter := zip.NewWriter(&zipBuffer)

		csvFile, _ := zipWriter.Create("data.csv")
		csvFile.Write(csvBuffer.Bytes())

		if err := zipWriter.Close(); err != nil {
			log.Printf("Zip close error: %v", err)
			http.Error(w, "Archive error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.Write(zipBuffer.Bytes())
	}
}
