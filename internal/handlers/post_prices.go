package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// HandlerPostPrices обрабатывает POST-запрос для загрузки цен
func HandlerPostPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling POST request for /api/v0/prices")

		// Парсинг multipart/form-data запроса
		err := r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to get file from form", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Логирование имени файла
		fmt.Printf("Received file: %s\n", header.Filename)

		// Чтение содержимого файла в память
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Unable to read file content", http.StatusInternalServerError)
			return
		}

		// Создание reader для zip-файла
		zipReader, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
		if err != nil {
			http.Error(w, "Unable to read zip file", http.StatusInternalServerError)
			return
		}

		// Переменные для подсчета статистики
		var (
			totalItems       int
			totalPrice       float64
			uniqueCategories = make(map[string]struct{}) // Для хранения уникальных категорий
		)

		// Обработка каждого файла в zip-архиве
		for _, zipFile := range zipReader.File {
			// Пропуск файлов, которые не являются CSV
			if !strings.HasSuffix(zipFile.Name, ".csv") {
				fmt.Printf("Skipping non-CSV file: %s\n", zipFile.Name)
				continue
			}

			fmt.Printf("Processing CSV file: %s\n", zipFile.Name)

			// Открытие файла в zip-архиве
			fileInZip, err := zipFile.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to open file %s in zip", zipFile.Name), http.StatusInternalServerError)
				return
			}
			defer fileInZip.Close()

			// Чтение CSV-файла
			csvReader := csv.NewReader(fileInZip)
			records, err := csvReader.ReadAll()
			if err != nil {
				http.Error(w, fmt.Sprintf("Unable to read CSV file %s", zipFile.Name), http.StatusInternalServerError)
				return
			}

			// Пропуск заголовка и обработка каждой строки
			for i, record := range records {
				if i == 0 { // Пропуск первой строки (заголовок)
					continue
				}

				// Парсинг данных из CSV
				productID, _ := strconv.Atoi(record[0])       // ID продукта
				productName := record[1]                      // Название продукта
				category := record[2]                         // Категория продукта
				price, _ := strconv.ParseFloat(record[3], 64) // Цена продукта
				createdAt := record[4]                        // Дата создания

				// Запись данных в PostgreSQL
				_, err := pool.Exec(context.Background(),
					"INSERT INTO prices (product_id, product_name, category, price, created_at) VALUES ($1, $2, $3, $4, $5)",
					productID, productName, category, price, createdAt,
				)
				if err != nil {
					fmt.Printf("Failed to insert record %d: %v\n", i, err)
					continue
				}

				// Обновление статистики
				totalItems++
				totalPrice += price
				uniqueCategories[category] = struct{}{} // Добавление категории в множество
			}
		}

		// Подготовка JSON-ответа
		response := map[string]interface{}{
			"total_items":      totalItems,
			"total_categories": len(uniqueCategories), // Количество уникальных категорий
			"total_price":      int(totalPrice),       // Округляем totalPrice до целого числа
		}

		// Преобразование ответа в JSON
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Unable to encode JSON response", http.StatusInternalServerError)
			return
		}

		// Установка заголовков и отправка JSON-ответа
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)

		// Логирование успешного завершения
		fmt.Println("Successfully processed file and returned JSON response")
	}
}
