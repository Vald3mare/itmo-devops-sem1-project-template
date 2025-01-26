package handlers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"project_sem/internal/db"
	"strconv"

	"context"
	"log"
)

type PostResponse struct {
	TotalItems     int     `json:"total_items"`
	TotalCategories int    `json:"total_categories"`
	TotalPrice     float64 `json:"total_price"`
}

func PostPrices(w http.ResponseWriter, r *http.Request) {
	// Получение файла из запроса
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Сохранение архива во временный файл
	tempFile, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	io.Copy(tempFile, file)
	tempFile.Close()

	// Разархивируем данные
	archive, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, "Failed to unzip file", http.StatusInternalServerError)
		return
	}
	defer archive.Close()

	var totalItems int
	var totalPrice float64
	categorySet := make(map[string]struct{})

	// Обрабатываем файл data.csv
	for _, file := range archive.File {
		if filepath.Base(file.Name) == "data.csv" {
			f, err := file.Open()
			if err != nil {
				http.Error(w, "Failed to open CSV file", http.StatusInternalServerError)
				return
			}
			defer f.Close()

			// Читаем CSV
			reader := csv.NewReader(f)
			for {
				record, err := reader.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					http.Error(w, "Failed to read CSV file", http.StatusInternalServerError)
					return
				}

				// Парсим данные
				id, _ := strconv.Atoi(record[0])
				name := record[1]
				category := record[2]
				price, _ := strconv.ParseFloat(record[3], 64)
				createDate := record[4]

				// Вставляем в базу данных
				query := `INSERT INTO prices (id, name, category, price, create_date)
                          VALUES ($1, $2, $3, $4, $5)`
				_, err = db.DBPool.Exec(context.Background(), query, id, name, category, price, createDate)
				if err != nil {
					log.Println("Failed to insert row:", err)
					continue
				}

				// Обновляем метрики
				totalItems++
				totalPrice += price
				categorySet[category] = struct{}{}
			}
		}
	}

	// Формируем ответ
	response := PostResponse{
		TotalItems:     totalItems,
		TotalCategories: len(categorySet),
		TotalPrice:     totalPrice,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
