package handlers

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var db *sql.DB

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Запрашиваем все данные из БД
		rows, err := db.Query(`SELECT id, product_name, category, price, creation_date FROM prices`)
		if err != nil {
			http.Error(w, "Failed to query database", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Сохраняем данные в память
		type Product struct {
			ID           int
			Name         string
			Category     string
			Price        float64
			CreationDate string
		}
		var products []Product

		for rows.Next() {
			var p Product
			err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.CreationDate)
			if err != nil {
				log.Println("Error: Failed to scan row:", err)
				http.Error(w, "Failed to scan row", http.StatusInternalServerError)
				return
			}
			products = append(products, p)
		}

		if err = rows.Err(); err != nil {
			log.Println("Error: Error occurred while iterating over rows:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		log.Printf("Loaded %d products from database", len(products))

		// Создаем CSV в памяти
		var csvBuffer bytes.Buffer
		writer := csv.NewWriter(&csvBuffer)

		// Записываем заголовок
		err = writer.Write([]string{"id", "name", "category", "price", "create_date"})
		if err != nil {
			log.Println("Error: Failed to write CSV header:", err)
			http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
			return
		}

		// Записываем строки
		for _, p := range products {
			err = writer.Write([]string{
				strconv.Itoa(p.ID),
				p.Name,
				p.Category,
				fmt.Sprintf("%.2f", p.Price),
				p.CreationDate,
			})
			if err != nil {
				log.Println("Error: Failed to write CSV row:", err)
				http.Error(w, "Failed to write CSV row", http.StatusInternalServerError)
				return
			}
		}
		writer.Flush()

		if err := writer.Error(); err != nil {
			log.Println("Error: CSV writing error:", err)
			http.Error(w, "CSV writing error", http.StatusInternalServerError)
			return
		}

		log.Println("CSV file created in memory")

		// Создаем ZIP-архив в памяти
		var zipBuffer bytes.Buffer
		zipWriter := zip.NewWriter(&zipBuffer)

		csvFile, err := zipWriter.Create("data.csv")
		if err != nil {
			log.Println("Error: Failed to create CSV inside ZIP:", err)
			http.Error(w, "Failed to create CSV inside ZIP", http.StatusInternalServerError)
			return
		}

		_, err = csvFile.Write(csvBuffer.Bytes())
		if err != nil {
			log.Println("Error: Failed to write CSV to ZIP:", err)
			http.Error(w, "Failed to write CSV to ZIP", http.StatusInternalServerError)
			return
		}

		err = zipWriter.Close()
		if err != nil {
			log.Println("Error: Failed to close ZIP archive:", err)
			http.Error(w, "Failed to close ZIP archive", http.StatusInternalServerError)
			return
		}

		log.Println("ZIP archive created in memory")

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.Header().Set("Content-Length", strconv.Itoa(zipBuffer.Len()))
		_, err = w.Write(zipBuffer.Bytes())

		if err != nil {
			log.Println("Error: Failed to send ZIP file:", err)
			http.Error(w, "Failed to send ZIP file", http.StatusInternalServerError)
			return
		}

		log.Println("ZIP file sent successfully")
	}
}
