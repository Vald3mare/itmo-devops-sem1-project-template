package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strconv"
	"time"
)

// HandlerGetPrices обрабатывает GET-запрос для получения данных из базы данных
func HandlerGetPrices(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := pool.Query(context.Background(), "SELECT product_id, product_name, category, price, created_at FROM prices")
		if err != nil {
			http.Error(w, "Unable to query database: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var csvData bytes.Buffer
		csvWriter := csv.NewWriter(&csvData)

		csvWriter.Write([]string{"id", "name", "category", "price", "create_date"})

		for rows.Next() {
			var (
				productID   int
				productName string
				category    string
				price       float64
				createdAt   time.Time
			)

			err := rows.Scan(&productID, &productName, &category, &price, &createdAt)
			if err != nil {
				http.Error(w, "Не удалось прочитать строку из БД: "+err.Error(), http.StatusInternalServerError)
				return
			}

			record := []string{
				strconv.Itoa(productID),
				productName,
				category,
				strconv.FormatFloat(price, 'f', 2, 64),
				createdAt.Format("2006-01-02"),
			}

			csvWriter.Write(record)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Ошибка после сканирования строк из БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			http.Error(w, "Не удалось записать данные в CSV: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var zipData bytes.Buffer
		zipWriter := zip.NewWriter(&zipData)

		fileWriter, err := zipWriter.Create("data.csv")
		if err != nil {
			http.Error(w, "Не удалось создать файл внутри zip-архива: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = fileWriter.Write(csvData.Bytes())
		if err != nil {
			http.Error(w, "Не удалось записать CSV данные внутрь zip-архива: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = zipWriter.Close()
		if err != nil {
			http.Error(w, "Не удалось закрыть zip-архив: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename=data.zip")
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Length", strconv.Itoa(zipData.Len()))

		w.WriteHeader(http.StatusOK)
		w.Write(zipData.Bytes())
	}
}
