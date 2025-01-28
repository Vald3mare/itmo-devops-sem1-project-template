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
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "ошибка чтения файла из запроса", http.StatusBadRequest)
			return
		}
		defer file.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			http.Error(w, "не удалось записать файл в буфер", http.StatusInternalServerError)
			return
		}
		fileData := buf.Bytes()

		zipReader, err := zip.NewReader(bytes.NewReader(fileData), int64(len(fileData)))
		if err != nil {
			http.Error(w, "не удалось разархивировать файл", http.StatusBadRequest)
			return
		}

		var csvFile *zip.File
		for _, f := range zipReader.File {
			if strings.HasSuffix(f.Name, "data.csv") {
				csvFile = f
				break
			}
		}
		if csvFile == nil {
			http.Error(w, "файл data.csv не был найден в архиве", http.StatusBadRequest)
			return
		}

		fileReader, err := csvFile.Open()
		if err != nil {
			http.Error(w, "ошибка чтения файла data.csv", http.StatusInternalServerError)
			return
		}
		defer fileReader.Close()

		lines, err := csv.NewReader(fileReader).ReadAll()
		if err != nil {
			http.Error(w, "ошибка парсинга данных из csv файла", http.StatusInternalServerError)
			return
		}

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
				http.Error(w, "некорретный формат цены", http.StatusBadRequest)
				return
			}

			_, err = db.Exec(context.Background(),
				`INSERT INTO prices (id, name, category, price, create_date)
				 VALUES ($1, $2, $3, $4, $5)
				 ON CONFLICT (id) DO UPDATE
				 SET name = EXCLUDED.name, category = EXCLUDED.category, price = EXCLUDED.price, create_date = EXCLUDED.create_date`,
				id, name, category, price, createDate)
			if err != nil {
				http.Error(w, "не удалось вставить данные в БД", http.StatusInternalServerError)
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
