package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// HandlerPostPrices обрабатывает POST-запрос для загрузки данных
func HandlerPostPrices(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Чтение zip-архива из тела запроса
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Считывание файла в память
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			http.Error(w, "Failed to read file into buffer", http.StatusInternalServerError)
			return
		}
		fileData := buf.Bytes()

		// Создание zip-ридера
		zipReader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
		if err != nil {
			http.Error(w, "Failed to unzip file", http.StatusBadRequest)
			return
		}

		// Поиск data.csv
		var csvFile *zip.File
		for _, f := range zipReader.File {
			if strings.HasSuffix(f.Name, "data.csv") {
				csvFile = f
				break
			}
		}
		if csvFile == nil {
			http.Error(w, "File 'data.csv' not found in archive", http.StatusBadRequest)
			return
		}

		// Чтение содержимого CSV
		fileReader, err := csvFile.Open()
		if err != nil {
			http.Error(w, "Failed to open CSV file", http.StatusInternalServerError)
			return
		}
		defer fileReader.Close()

		lines, err := csv.NewReader(fileReader).ReadAll()
		if err != nil {
			http.Error(w, "Failed to parse CSV file", http.StatusInternalServerError)
			return
		}

		// Обработка данных
		var totalItems int
		var totalPrice float64
		categories := make(map[string]struct{})

		for i, line := range lines {
			if i == 0 {
				continue // Пропускаем заголовок
			}

			id := line[0]
			name := line[1]
			category := line[2]
			priceStr := line[3]
			createDate := line[4]

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				http.Error(w, "Invalid price format", http.StatusBadRequest)
				return
			}

			_, err = db.Exec(context.Background(),
				`INSERT INTO prices (id, name, category, price, create_date)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE
				 SET name = EXCLUDED.name, category = EXCLUDED.category, price = EXCLUDED.price, create_date = EXCLUDED.create_date`,
				id, name, category, price, createDate)
			if err != nil {
				http.Error(w, "Failed to process database data", http.StatusInternalServerError)
				return
			}

			totalItems++
			categories[category] = struct{}{}
			totalPrice += price
		}

		// Формирование ответа
		response := map[string]interface{}{
			"total_items":      totalItems,
			"total_categories": len(categories),
			"total_price":      totalPrice,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
