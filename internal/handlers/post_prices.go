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
		log.Println("Received POST request to /api/v0/prices")

		// Читаем zip-файл из запроса
		file, _, err := r.FormFile("file")
		if err != nil {
			log.Println("Error: Failed to read uploaded file:", err)
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Читаем содержимое ZIP в память
		zipData, err := io.ReadAll(file)
		if err != nil {
			log.Println("Error: Failed to read ZIP file into memory:", err)
			http.Error(w, "Failed to read ZIP file", http.StatusInternalServerError)
			return
		}

		// Открываем ZIP-архив из памяти
		zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
		if err != nil {
			log.Println("Error: Failed to open ZIP file:", err)
			http.Error(w, "Failed to open ZIP file", http.StatusInternalServerError)
			return
		}

		var csvFile io.ReadCloser
		var foundCSV string

		// Ищем CSV-файл внутри архива
		for _, zipFile := range zipReader.File {
			if zipFile.FileInfo().IsDir() {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(zipFile.Name), ".csv") {
				continue
			}

			log.Println("Found CSV file inside ZIP:", zipFile.Name)

			csvFile, err = zipFile.Open()
			if err != nil {
				log.Println("Error: Failed to open CSV file:", err)
				http.Error(w, "Failed to open CSV file", http.StatusInternalServerError)
				return
			}

			log.Println("Successfully extracted CSV file:", csvFile)
			foundCSV = zipFile.Name

			break
		}

		if csvFile == nil {

			log.Println("Error: No CSV file found in archive")
			http.Error(w, "No CSV file found in archive", http.StatusBadRequest)
			return
		}

		log.Println("Extracted CSV file:", foundCSV)

		var csvBuffer bytes.Buffer
		_, err = io.Copy(&csvBuffer, csvFile)
		if err != nil {
			log.Println("❌ ERROR: Failed to copy CSV file to memory:", err)
			http.Error(w, "Failed to process CSV file", http.StatusInternalServerError)
			return
		}

		csvFile.Close()

		reader := csv.NewReader(bytes.NewReader(csvBuffer.Bytes()))
		reader.Comma = ','

		rows, err := reader.ReadAll()
		if err != nil {
			log.Println("Error: Failed to parse CSV file:", err)
			http.Error(w, "Failed to parse CSV file", http.StatusInternalServerError)
			return
		}

		log.Printf("CSV file contains %d rows (including header)", len(rows))

		// Проверяем данные перед вставкой в БД
		products := []struct {
			ProductID    int
			ProductName  string
			Category     string
			Price        float64
			CreationDate string
		}{}

		for i, row := range rows {
			// Пропускаем заголовок
			if i == 0 {
				continue
			}

			if len(row) != 5 {
				log.Printf("Skipping malformed row %d: %v", i, row)
				continue
			}

			productID, err := strconv.Atoi(row[0])
			if err != nil {
				log.Printf("Skipping row %d due to invalid product ID: %v", i, row[0])
				continue
			}

			price, err := strconv.ParseFloat(row[3], 64)
			if err != nil {
				log.Printf("Skipping row %d due to invalid price: %v", i, row[3])
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
				ProductName:  row[1],
				Category:     row[2],
				Price:        price,
				CreationDate: row[4],
			})
		}

		if len(products) == 0 {
			log.Println("Error: No valid data found in CSV")
			http.Error(w, "No valid data in CSV", http.StatusBadRequest)
			return
		}

		// Начинаем транзакцию для массовой вставки
		tx, err := db.Begin()
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
