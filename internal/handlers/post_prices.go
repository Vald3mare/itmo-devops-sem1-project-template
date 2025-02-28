package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"database/sql"
)

var database *sql.DB

func HandlerPostPrices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to parse file from request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		zipData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read ZIP file", http.StatusInternalServerError)
			return
		}

		zipReader, err := zip.NewReader(bytes.NewReader(zipData), header.Size)
		if err != nil {
			http.Error(w, "Failed to open ZIP file", http.StatusInternalServerError)
			return
		}

		var csvFile *zip.File
		for _, zipFile := range zipReader.File {
			if !strings.HasSuffix(strings.ToLower(zipFile.Name), ".csv") {
				continue
			}

			if !zipFile.FileInfo().IsDir() && strings.HasSuffix(strings.ToLower(zipFile.Name), ".csv") {
				csvFile = zipFile
				break
			}

			if csvFile == nil {
				http.Error(w, "CSV file not found in the archive", http.StatusBadRequest)
				return
			}

			csvReader, err := csvFile.Open()
			if err != nil {
				http.Error(w, "Unable to open CSV file", http.StatusInternalServerError)
				return
			}
			defer csvReader.Close()

			csvData := csv.NewReader(csvReader)
			csvData.Comma = ','

			records, err := csvData.ReadAll()
			if err != nil {
				http.Error(w, "Unable to parse CSV file", http.StatusInternalServerError)
				return
			}

			products := []struct {
				ProductID    int
				ProductName  string
				Category     string
				Price        float64
				CreationDate string
			}{}

			for i, record := range records {
				if i == 0 {
					continue
				}
				if len(record) != 5 {
					log.Printf("Skipping malformed row %d: %v", i, record)
					continue
				}

				productID, err := strconv.Atoi(record[0])
				if err != nil {
					log.Printf("Skipping row %d due to invalid product ID: %v", i, record[0])
					continue
				}

				price, err := strconv.ParseFloat(record[3], 64)
				if err != nil {
					log.Printf("Skipping row %d due to invalid price: %v", i, record[3])
					continue
				}

				products = append(products, struct {
					ProductID    int
					ProductName  string
					Category     string
					Price        float64
					CreationDate string
				}{
					ProductID:    productID,
					ProductName:  record[1],
					Category:     record[2],
					Price:        price,
					CreationDate: record[4],
				})

				if len(products) == 0 {
					log.Println("Error: No valid data found in CSV")
					http.Error(w, "No valid data in CSV", http.StatusBadRequest)
					return
				}

				tx, err := database.Begin()
				if err != nil {
					log.Println("Error: Failed to start transaction:", err)
					http.Error(w, "Database transaction error", http.StatusInternalServerError)
					return
				}

				stmt, err := tx.Prepare(`INSERT INTO prices (product_name, category, price, creation_date) VALUES ($1, $2, $3, $4)`)
				if err != nil {
					log.Println("Error: Failed to prepare statement:", err)
					tx.Rollback()
					http.Error(w, "Database preparation error", http.StatusInternalServerError)
					return
				}
				defer stmt.Close()

				totalItems := 0
				for _, p := range products {
					_, err = stmt.Exec(p.ProductName, p.Category, p.Price, p.CreationDate)
					if err != nil {
						log.Printf("Skipping row with Product ID %d due to database error: %v", p.ProductID, err)
						continue
					}
					totalItems++
					//totalCategories[p.Category] = struct{}{}
					//totalPrice += p.Price
				}

				err = tx.Commit()
				if err != nil {
					log.Println("Error: Failed to commit transaction:", err)
					http.Error(w, "Database commit error", http.StatusInternalServerError)
					return
				}

				log.Printf("Successfully inserted %d items into database", totalItems)

				// Запрашиваем статистику из БД
				var totalCategories int
				var totalPrice float64
				err = db.QueryRow(`SELECT COUNT(DISTINCT category), COALESCE(SUM(price), 0) FROM prices`).Scan(&totalCategories, &totalPrice)
				if err != nil {
					log.Println("Error: Failed to calculate statistics:", err)
					http.Error(w, "Failed to calculate statistics", http.StatusInternalServerError)
					return
				}

				response := map[string]interface{}{
					"total_items":      totalItems,
					"total_categories": totalCategories,
					"total_price":      totalPrice,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				log.Println("Response sent successfully")
			}
		}
	}
}
