package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"strconv"
	"time"
)

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлечение данных из базы данных
		rows, err := pool.Query(context.Background(), "SELECT product_id, product_name, category, price, created_at FROM prices")
		if err != nil {
			http.Error(w, "Unable to query database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Создание CSV-файла в памяти
		var csvData bytes.Buffer
		csvWriter := csv.NewWriter(&csvData)

		// Запись заголовка CSV
		csvWriter.Write([]string{"id", "name", "category", "price", "create_date"})

		// Запись данных в CSV
		for rows.Next() {
			var (
				productID   int
				productName string
				category    string
				price       float64
				createdAt   time.Time
			)

			// Сканирование данных из строки
			err := rows.Scan(&productID, &productName, &category, &price, &createdAt)
			if err != nil {
				http.Error(w, "Unable to scan row: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Преобразование данных в строки
			record := []string{
				strconv.Itoa(productID),
				productName,
				category,
				strconv.FormatFloat(price, 'f', 2, 64),
				createdAt.Format("2006-01-02"),
			}

			// Запись строки в CSV
			csvWriter.Write(record)
		}

		// Проверка на ошибки после завершения итерации
		if err := rows.Err(); err != nil {
			http.Error(w, "Error after scanning rows: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Завершение записи CSV
		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			http.Error(w, "Unable to write CSV data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Создание zip-архива в памяти
		var zipData bytes.Buffer
		zipWriter := zip.NewWriter(&zipData)

		// Добавление CSV-файла в архив
		fileWriter, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "Unable to create file in zip archive: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Запись данных CSV в файл внутри архива
		_, err = fileWriter.Write(csvData.Bytes())
		if err != nil {
			http.Error(w, "Unable to write CSV data to zip archive: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Завершение создания zip-архива
		err = zipWriter.Close()
		if err != nil {
			http.Error(w, "Unable to close zip archive: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Установка заголовков для скачивания файла
		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", strconv.Itoa(zipData.Len())) // Указываем размер файла

		// Отправка zip-архива клиенту
		w.WriteHeader(http.StatusOK)
		w.Write(zipData.Bytes())

		log.Println("Successfully sent zip archive to client")
	}
}
