package handlers

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Item представляет структуру данных для элемента из CSV-файла
type Item struct {
	Name       string
	Category   string
	Price      float64
	CreateDate time.Time
}

// Response представляет структуру JSON-ответа
type Response struct {
	TotalItems      int     `json:"total_items"`
	TotalCategories int     `json:"total_categories"`
	TotalPrice      float64 `json:"total_price"`
}

func HandlerPostPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ограничиваем размер файла (например, 10 МБ)
		r.ParseMultipartForm(10 << 20)

		// Получаем файл из запроса
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Unable to read file from request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Проверяем, что файл имеет расширение .zip
		if header.Header.Get("Content-Type") != "application/zip" {
			http.Error(w, "File must be a ZIP archive", http.StatusBadRequest)
			return
		}

		// Читаем ZIP-архив
		zipReader, err := zip.NewReader(file, header.Size)
		if err != nil {
			http.Error(w, "Unable to read ZIP archive", http.StatusInternalServerError)
			return
		}

		// Ищем CSV-файл в архиве
		var csvFile *zip.File
		for _, f := range zipReader.File {
			if !f.FileInfo().IsDir() && f.FileInfo().Name() == "data.csv" {
				csvFile = f
				break
			}
		}
		if csvFile == nil {
			http.Error(w, "CSV file not found in the archive", http.StatusBadRequest)
			return
		}

		// Открываем CSV-файл
		csvReader, err := csvFile.Open()
		if err != nil {
			http.Error(w, "Unable to open CSV file", http.StatusInternalServerError)
			return
		}
		defer csvReader.Close()

		// Парсим CSV-файл
		csvData := csv.NewReader(csvReader)
		records, err := csvData.ReadAll()
		if err != nil {
			http.Error(w, "Unable to parse CSV file", http.StatusInternalServerError)
			return
		}

		// Подготавливаем данные для вставки в базу данных
		var items []Item
		categories := make(map[string]bool)
		var totalPrice float64

		for i, record := range records {
			// Пропускаем заголовок (первую строку)
			if i == 0 {
				continue
			}

			// Парсим данные из CSV
			price, err := strconv.ParseFloat(record[3], 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid price in row %d", i+1), http.StatusBadRequest)
				return
			}

			createDate, err := time.Parse("2006-01-02", record[4])
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid create_date in row %d", i+1), http.StatusBadRequest)
				return
			}

			item := Item{
				Name:       record[1],
				Category:   record[2],
				Price:      price,
				CreateDate: createDate,
			}
			items = append(items, item)

			// Считаем уникальные категории
			categories[item.Category] = true

			// Считаем общую стоимость
			totalPrice += item.Price
		}

		// Вставляем данные в базу данных
		for _, item := range items {
			_, err := pool.Exec(r.Context(), `
				INSERT INTO prices (product_name, category, price, creation_date)
				VALUES ($1, $2, $3, $4)
			`, item.Name, item.Category, item.Price, item.CreateDate)
			if err != nil {
				http.Error(w, "Unable to insert data into database", http.StatusInternalServerError)
				return
			}
		}

		// Формируем JSON-ответ
		response := Response{
			TotalItems:      len(items),
			TotalCategories: len(categories),
			TotalPrice:      totalPrice,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
